# Pilot Control Telemetry Gateway

Pilot Control provides an [Open Telemetry](https://opentelemetry.io/) Gateway to ingest metrics sent by Pilot agents.

The gateway uses the [Open Telemetry Protocol](https://opentelemetry.io/docs/reference/specification/protocol/otlp/) to encode, transport, and deliver telemetry data between pilot agents and send it to a designated backend.

Only one processing backend can be configured at any point in time. Pilot Control uses a connector to remain agnostic of the backend where data is to be sent.

The following figure shows the process of collecting and exporting telemetry:

![ot-gateway](ot-gateway.drawio.png)

An [Open Telemetry collector](https://github.com/open-telemetry/opentelemetry-collector) collects telemetry data and it in a persistent file queue. 
This is done to prevent losing telemetry information if the device or the network go down.

The pilot agent in the device inspect the queue and exports the metrics to Pilot Control's Open Telemetry gateway.

Pilot control hands the information over to a connector that in turn transfers the telemetry to a designated streaming backend.