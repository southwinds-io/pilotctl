/*
   Pilot Control Service
   Copyright (C) 2022-Present SouthWinds Tech Ltd - www.southwinds.io

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"strconv"
	"time"
)

const (
	MetricHistogramBoundKeyV2              = "le"
	MetricHistogramCountSuffix             = "_count"
	MetricHistogramSumSuffix               = "_sum"
	MetricHistogramBucketSuffix            = "_bucket"
	MetricSummaryQuantileKeyV2             = "quantile"
	MetricSummaryCountSuffix               = "_count"
	MetricSummarySumSuffix                 = "_sum"
	AttributeInstrumentationLibraryName    = "otel.library.name"
	AttributeInstrumentationLibraryVersion = "otel.library.version"
)

type OtelDataPointConverter struct {
	logger Logger
	points []DataPoint
}

func NewOtelDataPointConverter() OtelDataPointConverter {
	return OtelDataPointConverter{
		logger: warningLogger{},
		points: make([]DataPoint, 0),
	}
}

type Logger interface {
	Debug(msg string, kv ...interface{})
}

func (c *OtelDataPointConverter) Convert(md pmetric.Metrics) ([]DataPoint, error) {
	var err error
	c.points = make([]DataPoint, 0)
	for i := 0; i < md.ResourceMetrics().Len(); i++ {
		resourceMetrics := md.ResourceMetrics().At(i)
		for j := 0; j < resourceMetrics.ScopeMetrics().Len(); j++ {
			ilMetrics := resourceMetrics.ScopeMetrics().At(j)
			for k := 0; k < ilMetrics.Metrics().Len(); k++ {
				metric := ilMetrics.Metrics().At(k)
				if err = c.writeMetric(resourceMetrics.Resource(), ilMetrics.Scope(), metric); err != nil {
					return nil, fmt.Errorf("failed to convert OTLP metric to data point: %s", err)
				}
			}
		}
	}
	return c.points, nil
}

func (c *OtelDataPointConverter) writePoint(measurement string, tags map[string]string, fields map[string]interface{}, ts time.Time, vType MetricValueType) error {
	c.points = append(c.points, DataPoint{
		Measure: measurement,
		Tags:    tags,
		Fields:  fields,
		Time:    ts,
		Type:    vType.String(),
	})
	return nil
}

func (c *OtelDataPointConverter) writeMetric(resource pcommon.Resource, instrumentationLibrary pcommon.InstrumentationScope, metric pmetric.Metric) error {
	// Ignore metric.Description() and metric.Unit() .
	switch metric.Type() {
	case pmetric.MetricTypeGauge:
		return c.writeGauge(resource, instrumentationLibrary, metric.Name(), metric.Gauge())
	case pmetric.MetricTypeSum:
		if metric.Sum().IsMonotonic() {
			return c.writeSum(resource, instrumentationLibrary, metric.Name(), metric.Sum())
		}
		return c.writeGaugeFromSum(resource, instrumentationLibrary, metric.Name(), metric.Sum())
	case pmetric.MetricTypeHistogram:
		return c.writeHistogram(resource, instrumentationLibrary, metric.Name(), metric.Histogram())
	case pmetric.MetricTypeSummary:
		return c.writeSummary(resource, instrumentationLibrary, metric.Name(), metric.Summary())
	default:
		return fmt.Errorf("unknown metric type %q", metric.Type())
	}
}

func (c *OtelDataPointConverter) writeGauge(resource pcommon.Resource, instrumentationLibrary pcommon.InstrumentationScope, measurement string, gauge pmetric.Gauge) error {
	for i := 0; i < gauge.DataPoints().Len(); i++ {
		dataPoint := gauge.DataPoints().At(i)
		tags, fields, ts, err := c.initMetricTagsAndTimestamp(resource, instrumentationLibrary, dataPoint.Timestamp(), dataPoint.Attributes())
		if err != nil {
			return err
		}

		switch dataPoint.ValueType() {
		case pmetric.NumberDataPointValueTypeEmpty:
			continue
		case pmetric.NumberDataPointValueTypeDouble:
			fields[measurement] = dataPoint.DoubleValue()
		case pmetric.NumberDataPointValueTypeInt:
			fields[measurement] = dataPoint.IntValue()
		default:
			return fmt.Errorf("unsupported gauge data point type %d", dataPoint.ValueType())
		}
		if err = c.writePoint(measurement, tags, fields, ts, MetricValueTypeGauge); err != nil {
			return fmt.Errorf("failed to write point for gauge: %w", err)
		}
	}
	return nil
}

func (c *OtelDataPointConverter) writeSum(resource pcommon.Resource, instrumentationLibrary pcommon.InstrumentationScope, measurement string, sum pmetric.Sum) error {
	if sum.AggregationTemporality() != pmetric.AggregationTemporalityCumulative {
		return fmt.Errorf("unsupported sum aggregation temporality %q", sum.AggregationTemporality())
	}
	for i := 0; i < sum.DataPoints().Len(); i++ {
		dataPoint := sum.DataPoints().At(i)
		tags, fields, ts, err := c.initMetricTagsAndTimestamp(resource, instrumentationLibrary, dataPoint.Timestamp(), dataPoint.Attributes())
		if err != nil {
			return err
		}

		switch dataPoint.ValueType() {
		case pmetric.NumberDataPointValueTypeEmpty:
			continue
		case pmetric.NumberDataPointValueTypeDouble:
			fields[measurement] = dataPoint.DoubleValue()
		case pmetric.NumberDataPointValueTypeInt:
			fields[measurement] = dataPoint.IntValue()
		default:
			return fmt.Errorf("unsupported sum data point type %d", dataPoint.ValueType())
		}
		if err = c.writePoint(measurement, tags, fields, ts, MetricValueTypeSum); err != nil {
			return fmt.Errorf("failed to write point for sum: %w", err)
		}
	}
	return nil
}

func (c *OtelDataPointConverter) writeGaugeFromSum(resource pcommon.Resource, instrumentationLibrary pcommon.InstrumentationScope, measurement string, sum pmetric.Sum) error {
	if sum.AggregationTemporality() != pmetric.AggregationTemporalityCumulative {
		return fmt.Errorf("unsupported sum (as gauge) aggregation temporality %q", sum.AggregationTemporality())
	}
	for i := 0; i < sum.DataPoints().Len(); i++ {
		dataPoint := sum.DataPoints().At(i)
		tags, fields, ts, err := c.initMetricTagsAndTimestamp(resource, instrumentationLibrary, dataPoint.Timestamp(), dataPoint.Attributes())
		if err != nil {
			return err
		}
		switch dataPoint.ValueType() {
		case pmetric.NumberDataPointValueTypeEmpty:
			continue
		case pmetric.NumberDataPointValueTypeDouble:
			fields[measurement] = dataPoint.DoubleValue()
		case pmetric.NumberDataPointValueTypeInt:
			fields[measurement] = dataPoint.IntValue()
		default:
			return fmt.Errorf("unsupported sum (as gauge) data point type %d", dataPoint.ValueType())
		}
		if err = c.writePoint(measurement, tags, fields, ts, MetricValueTypeGauge); err != nil {
			return fmt.Errorf("failed to write point for sum (as gauge): %w", err)
		}
	}
	return nil
}

func (c *OtelDataPointConverter) writeHistogram(resource pcommon.Resource, instrumentationLibrary pcommon.InstrumentationScope, measurement string, histogram pmetric.Histogram) error {
	if histogram.AggregationTemporality() != pmetric.AggregationTemporalityCumulative {
		return fmt.Errorf("unsupported histogram aggregation temporality %q", histogram.AggregationTemporality())
	}
	for i := 0; i < histogram.DataPoints().Len(); i++ {
		dataPoint := histogram.DataPoints().At(i)
		tags, fields, ts, err := c.initMetricTagsAndTimestamp(resource, instrumentationLibrary, dataPoint.Timestamp(), dataPoint.Attributes())
		if err != nil {
			return err
		}
		{
			f := make(map[string]interface{}, len(fields)+2)
			for k, v := range fields {
				f[k] = v
			}

			f[measurement+MetricHistogramCountSuffix] = float64(dataPoint.Count())
			f[measurement+MetricHistogramSumSuffix] = dataPoint.Sum()

			if err = c.writePoint(measurement, tags, f, ts, MetricValueTypeHistogram); err != nil {
				return fmt.Errorf("failed to write point for histogram: %w", err)
			}
		}

		bucketCounts, explicitBounds := dataPoint.BucketCounts(), dataPoint.ExplicitBounds()
		if bucketCounts.Len() > 0 &&
			bucketCounts.Len() != explicitBounds.Len() &&
			bucketCounts.Len() != explicitBounds.Len()+1 {
			// The infinity bucket is not used in this schema,
			// so accept input if that particular bucket is missing.
			return fmt.Errorf("invalid metric histogram bucket counts qty %d vs explicit bounds qty %d", bucketCounts.Len(), explicitBounds.Len())
		}

		for i := 0; i < explicitBounds.Len(); i++ {
			t := make(map[string]string, len(tags)+1)
			for k, v := range tags {
				t[k] = v
			}
			f := make(map[string]interface{}, len(fields)+1)
			for k, v := range fields {
				f[k] = v
			}

			boundTagValue := strconv.FormatFloat(explicitBounds.At(i), 'f', -1, 64)
			t[MetricHistogramBoundKeyV2] = boundTagValue
			f[measurement+MetricHistogramBucketSuffix] = float64(bucketCounts.At(i))

			if err = c.writePoint(measurement, t, f, ts, MetricValueTypeHistogram); err != nil {
				return fmt.Errorf("failed to write point for histogram: %w", err)
			}
		}
	}
	return nil
}

func (c *OtelDataPointConverter) writeSummary(resource pcommon.Resource, instrumentationLibrary pcommon.InstrumentationScope, measurement string, summary pmetric.Summary) error {
	for i := 0; i < summary.DataPoints().Len(); i++ {
		dataPoint := summary.DataPoints().At(i)
		tags, fields, ts, err := c.initMetricTagsAndTimestamp(resource, instrumentationLibrary, dataPoint.Timestamp(), dataPoint.Attributes())
		if err != nil {
			return err
		}
		{
			f := make(map[string]interface{}, len(fields)+2)
			for k, v := range fields {
				f[k] = v
			}

			f[measurement+MetricSummaryCountSuffix] = float64(dataPoint.Count())
			f[measurement+MetricSummarySumSuffix] = dataPoint.Sum()

			if err = c.writePoint(measurement, tags, f, ts, MetricValueTypeSummary); err != nil {
				return fmt.Errorf("failed to write point for summary: %w", err)
			}
		}
		for j := 0; j < dataPoint.QuantileValues().Len(); j++ {
			valueAtQuantile := dataPoint.QuantileValues().At(j)
			t := make(map[string]string, len(tags)+1)
			for k, v := range tags {
				t[k] = v
			}
			f := make(map[string]interface{}, len(fields)+1)
			for k, v := range fields {
				f[k] = v
			}

			quantileTagValue := strconv.FormatFloat(valueAtQuantile.Quantile(), 'f', -1, 64)
			t[MetricSummaryQuantileKeyV2] = quantileTagValue
			f[measurement] = valueAtQuantile.Value()

			if err = c.writePoint(measurement, t, f, ts, MetricValueTypeSummary); err != nil {
				return fmt.Errorf("failed to write point for summary: %w", err)
			}
		}
	}
	return nil
}

func (c *OtelDataPointConverter) initMetricTagsAndTimestamp(resource pcommon.Resource, instrumentationLibrary pcommon.InstrumentationScope, timestamp pcommon.Timestamp, attributes pcommon.Map) (tags map[string]string, fields map[string]interface{}, ts time.Time, err error) {
	ts = timestamp.AsTime()
	if ts.IsZero() {
		err = errors.New("metric has no timestamp")
		return
	}

	tags = make(map[string]string)
	fields = make(map[string]interface{})

	attributes.Range(func(k string, v pcommon.Value) bool {
		if k == "" {
			// c.logger.Debug("metric attribute key is empty")
		} else {
			var vv string
			vv, err = attributeValueToInfluxTagValue(v)
			if err != nil {
				return false
			}
			tags[k] = vv
		}
		return true
	})
	if err != nil {
		err = fmt.Errorf("failed to convert attribute value to string: %w", err)
		return
	}

	tags = resourceToTags(c.logger, resource, tags)
	tags = instrumentationLibraryToTags(instrumentationLibrary, tags)

	return
}

func resourceToTags(logger Logger, resource pcommon.Resource, tags map[string]string) (tagsAgain map[string]string) {
	resource.Attributes().Range(func(k string, v pcommon.Value) bool {
		if k == "" {
			logger.Debug("resource attribute key is empty")
		} else if v, err := attributeValueToInfluxTagValue(v); err != nil {
			logger.Debug("invalid resource attribute value", "key", k, err)
		} else {
			tags[k] = v
		}
		return true
	})
	return tags
}

func instrumentationLibraryToTags(instrumentationLibrary pcommon.InstrumentationScope, tags map[string]string) (tagsAgain map[string]string) {
	if instrumentationLibrary.Name() != "" {
		tags[AttributeInstrumentationLibraryName] = instrumentationLibrary.Name()
	}
	if instrumentationLibrary.Version() != "" {
		tags[AttributeInstrumentationLibraryVersion] = instrumentationLibrary.Version()
	}
	return tags
}

func attributeValueToInfluxTagValue(value pcommon.Value) (string, error) {
	switch value.Type() {
	case pcommon.ValueTypeStr:
		return value.Str(), nil
	case pcommon.ValueTypeInt:
		return strconv.FormatInt(value.Int(), 10), nil
	case pcommon.ValueTypeDouble:
		return strconv.FormatFloat(value.Double(), 'f', -1, 64), nil
	case pcommon.ValueTypeBool:
		return strconv.FormatBool(value.Bool()), nil
	case pcommon.ValueTypeMap:
		if jsonBytes, err := json.Marshal(otlpKeyValueListToMap(value.Map())); err != nil {
			return "", err
		} else {
			return string(jsonBytes), nil
		}
	case pcommon.ValueTypeSlice:
		if jsonBytes, err := json.Marshal(otlpArrayToSlice(value.Slice())); err != nil {
			return "", err
		} else {
			return string(jsonBytes), nil
		}
	case pcommon.ValueTypeEmpty:
		return "", nil
	default:
		return "", fmt.Errorf("unknown value type %d", value.Type())
	}
}

func otlpKeyValueListToMap(kvList pcommon.Map) map[string]interface{} {
	m := make(map[string]interface{}, kvList.Len())
	kvList.Range(func(k string, v pcommon.Value) bool {
		switch v.Type() {
		case pcommon.ValueTypeStr:
			m[k] = v.Str()
		case pcommon.ValueTypeInt:
			m[k] = v.Int()
		case pcommon.ValueTypeDouble:
			m[k] = v.Double()
		case pcommon.ValueTypeBool:
			m[k] = v.Bool()
		case pcommon.ValueTypeMap:
			m[k] = otlpKeyValueListToMap(v.Map())
		case pcommon.ValueTypeSlice:
			m[k] = otlpArrayToSlice(v.Slice())
		case pcommon.ValueTypeEmpty:
			m[k] = nil
		default:
			m[k] = fmt.Sprintf("<invalid map value> %v", v)
		}
		return true
	})
	return m
}

func otlpArrayToSlice(arr pcommon.Slice) []interface{} {
	s := make([]interface{}, 0, arr.Len())
	for i := 0; i < arr.Len(); i++ {
		v := arr.At(i)
		switch v.Type() {
		case pcommon.ValueTypeStr:
			s = append(s, v.Str())
		case pcommon.ValueTypeInt:
			s = append(s, v.Int())
		case pcommon.ValueTypeDouble:
			s = append(s, v.Double())
		case pcommon.ValueTypeBool:
			s = append(s, v.Bool())
		case pcommon.ValueTypeEmpty:
			s = append(s, nil)
		default:
			s = append(s, fmt.Sprintf("<invalid array value> %v", v))
		}
	}
	return s
}

type DataPoint struct {
	Measure string
	Tags    map[string]string
	Fields  map[string]interface{}
	Time    time.Time
	Type    string
}

type MetricValueType uint8

const (
	MetricValueTypeUntyped MetricValueType = iota
	MetricValueTypeGauge
	MetricValueTypeSum
	MetricValueTypeHistogram
	MetricValueTypeSummary
)

func (vType MetricValueType) String() string {
	switch vType {
	case MetricValueTypeUntyped:
		return "untyped"
	case MetricValueTypeGauge:
		return "gauge"
	case MetricValueTypeSum:
		return "sum"
	case MetricValueTypeHistogram:
		return "histogram"
	case MetricValueTypeSummary:
		return "summary"
	default:
		panic("invalid InfluxMetricValueType")
	}
}

type warningLogger struct{}

func (warningLogger) Debug(msg string, args ...interface{}) {
	fmt.Printf("WARNING: %s\n", fmt.Sprintf(msg, args...))
}
