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
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"southwinds.dev/artisan/core"
	"southwinds.dev/artisan/data"
	"southwinds.dev/artisan/registry"
	h "southwinds.dev/http"
	"southwinds.dev/interlink-client"
	os2 "southwinds.dev/os"
	. "southwinds.dev/pilotctl/types"
	"strconv"
	"strings"
	"time"
)

// API backend services API
type API struct {
	conf  *Conf
	db    *Db
	iLink *ilink.Client
	// host information
	hostUUID string
	hostname string
	hostIP   string
}

func NewAPI(cfg *Conf) (*API, error) {
	db, err := NewDb(cfg.getDbHost(), cfg.getDbPort(), cfg.getDbName(), cfg.getDbUser(), cfg.getDbPwd(), cfg.getDbMaxConn())
	if err != nil {
		return nil, err
	}
	ilCfg := &ilink.ClientConf{
		BaseURI:            cfg.getOxWapiUrl(),
		Username:           cfg.getOxWapiUsername(),
		Password:           cfg.getOxWapiPassword(),
		InsecureSkipVerify: cfg.getOxWapiInsecureSkipVerify(),
	}
	ilCfg.SetAuthMode("basic")
	il, err := ilink.NewClient(ilCfg)
	if err != nil {
		return nil, fmt.Errorf("cannot create interlink http client: %s", err)
	}
	return &API{
		db:    db,
		conf:  cfg,
		iLink: il,
	}, nil
}

func (r *API) PingInterval() time.Duration {
	return r.conf.PingIntervalSecs()
}

func (r *API) Register(reg *RegistrationRequest) (*RegistrationResponse, error) {
	// registers the host with the cmdb
	result, err := r.iLink.PutItem(&ilink.Item{
		Key:         strings.ToUpper(fmt.Sprintf("HOST:%s", reg.MachineId)),
		Name:        strings.ToUpper(fmt.Sprintf("HOST_%s", strings.Replace(reg.Hostname, " ", "_", -1))),
		Description: "Pilot registered remote host",
		Status:      0,
		Type:        "U_HOST", // universal model host
		Tag:         nil,
		Meta:        nil,
		Txt:         "",
		// TODO: add mac address attribute
		Attribute: map[string]interface{}{
			"CPU":        reg.CPUs,
			"OS":         reg.OS,
			"MEMORY":     reg.TotalMemory,
			"PLATFORM":   reg.Platform,
			"VIRTUAL":    reg.Virtual,
			"IP":         reg.HostIP,
			"MACHINE-ID": reg.MachineId,
			"HOSTNAME":   reg.Hostname,
		},
	})
	// business error?
	if result != nil && result.Error {
		// return it
		return nil, fmt.Errorf(result.Message)
	}
	// otherwise, return technical error or nil if successful
	return &RegistrationResponse{
		Operation: result.Operation,
	}, err
}

func (r *API) GetCommandValue(fxKey string) (*CmdInfo, error) {
	item, err := r.iLink.GetItem(&ilink.Item{Key: fxKey})
	if err != nil {
		return nil, fmt.Errorf("cannot retrieve function specification from Onix: %s", err)
	}
	input, err := getInputFromMap(item.Meta)
	if err != nil {
		return nil, fmt.Errorf("cannot get input from map: %s", err)
	}
	return &CmdInfo{
		Function:      item.GetStringAttr("FX"),
		Package:       item.GetStringAttr("PACKAGE"),
		User:          item.GetStringAttr("USER"),
		Pwd:           item.GetStringAttr("PWD"),
		Verbose:       item.GetBoolAttr("VERBOSE"),
		Containerised: item.GetBoolAttr("CONTAINERISED"),
		Input:         input,
	}, nil
}

func (r *API) Ping() (jobId int64, fxKey string, fxVersion int64, err error) {
	rows, err := r.db.Query("select * from pilotctl_beat($1)", r.hostUUID)
	if err != nil {
		return -1, "", -1, err
	}
	for rows.Next() {
		err = rows.Scan(&jobId, &fxKey, &fxVersion)
		if err != nil {
			return 0, "", 0, err
		}
	}
	// returns the next job to execute for the machine Id or -1 if no job is available
	return jobId, fxKey, fxVersion, nil
}

// GetHosts get a list of hosts filtered by
// oGroup: organisation group key
// or: organisation key
// ar: area key
// loc: location key
func (r *API) GetHosts(oGroup, or, ar, loc string, label []string) ([]Host, error) {
	hosts := make([]Host, 0)
	pingInterval := r.PingInterval().Seconds()
	// note: the interval for determining if host is down is set to 2 pings
	rows, err := r.db.Query("select * from pilotctl_get_hosts($1, $2, $3, $4, $5, $6)", fmt.Sprintf("%.0f secs", 2*pingInterval), oGroup, or, ar, loc, label)
	if err != nil {
		return nil, fmt.Errorf("cannot get hosts: %s\n", err)
	}
	var (
		id            int64
		uuId          string
		macAddress    string
		connected     bool
		lastSeen      sql.NullTime
		orgGroup      sql.NullString
		org           sql.NullString
		area          sql.NullString
		location      sql.NullString
		inService     bool
		labels        []string
		scoreCritical sql.NullInt32
		scoreHigh     sql.NullInt32
		scoreMedium   sql.NullInt32
		scoreLow      sql.NullInt32
	)
	for rows.Next() {
		err = rows.Scan(&id, &uuId, &macAddress, &connected, &lastSeen, &orgGroup, &org, &area, &location, &inService, &labels, &scoreCritical, &scoreHigh, &scoreMedium, &scoreLow)
		if err != nil {
			return nil, err
		}
		var (
			tt        int64
			since     int
			sinceType string
		)
		if lastSeen.Valid {
			tt = lastSeen.Time.UnixNano()
			since, sinceType, err = toElapsedValues(lastSeen.Time.Format(time.RFC850))
			if err != nil {
				return nil, err
			}
		}
		hosts = append(hosts, Host{
			Id:             id,
			HostUUID:       uuId,
			HostMacAddress: macAddress,
			OrgGroup:       orgGroup.String,
			Org:            org.String,
			Area:           area.String,
			Location:       location.String,
			Connected:      connected,
			LastSeen:       tt,
			Since:          since,
			SinceType:      sinceType,
			Label:          labels,
			Critical:       int(scoreCritical.Int32),
			High:           int(scoreHigh.Int32),
			Medium:         int(scoreMedium.Int32),
			Low:            int(scoreLow.Int32),
		})
	}
	return hosts, rows.Err()
}

func (r *API) GetHostsAtLocations(locations []string) ([]Host, error) {
	hosts := make([]Host, 0)
	rows, err := r.db.Query("select * from pilotctl_get_host_at_location($1)", locations)
	if err != nil {
		return nil, fmt.Errorf("cannot get hosts: %s\n", err)
	}
	var (
		uuId       string
		macAddress string
		orgGroup   sql.NullString
		org        sql.NullString
		area       sql.NullString
		location   sql.NullString
	)
	for rows.Next() {
		err = rows.Scan(&uuId, &macAddress, &orgGroup, &org, &area, &location)
		if err != nil {
			return nil, err
		}
		hosts = append(hosts, Host{
			HostUUID:       uuId,
			HostMacAddress: macAddress,
			OrgGroup:       orgGroup.String,
			Org:            org.String,
			Area:           area.String,
			Location:       location.String,
		})
	}
	return hosts, nil
}

func (r *API) SetAdmission(admission Admission) error {
	if len(admission.HostUUID) == 0 {
		return fmt.Errorf("host UUID is missing")
	}
	return r.db.RunCommand("select pilotctl_set_admission($1, $2, $3, $4, $5, $6)",
		admission.HostUUID,
		admission.OrgGroup,
		admission.Org,
		admission.Area,
		admission.Location,
		admission.Label)
}

func (r *API) SetRegistration(registration Registration) error {
	if len(registration.MacAddress) == 0 {
		return fmt.Errorf("MAC-ADDRESS is missing")
	}
	return r.db.RunCommand("select pilotctl_set_registration($1, $2, $3, $4, $5, $6)",
		registration.MacAddress,
		registration.OrgGroup,
		registration.Org,
		registration.Area,
		registration.Location,
		registration.Label)
}

// AdmitRegistered admits a host that has been registered with a mac-address after confirmation of activation
func (r *API) AdmitRegistered(macAddress, hostUUID string) error {
	if len(macAddress) == 0 {
		return fmt.Errorf("mac-address is missing\n")
	}
	if len(hostUUID) == 0 {
		return fmt.Errorf("host UUID is missing\n")
	}
	return r.db.RunCommand("select pilotctl_admit_registered($1, $2)", macAddress, hostUUID)
}

// AuthenticatePilot authenticates pilot requests
func (r *API) AuthenticatePilot(token string) (*h.UserPrincipal, error) {
	if len(token) == 0 {
		msg := "authentication token is required and not provided"
		log.Println(msg)
		return nil, fmt.Errorf(msg)
	}
	value, err := base64.StdEncoding.DecodeString(reverse(token))
	if err != nil {
		msg := fmt.Sprintf("error decoding authentication token '%s': %s", token, err)
		log.Println(msg)
		return nil, fmt.Errorf(msg)
	}
	str := string(value)
	// token is: hostUUID (0) | hostIP (1) | hostName (2) | timestamp (3)
	parts := strings.Split(str, "|")
	tokenTime, err := strconv.ParseInt(parts[3], 10, 64)
	if err != nil {
		msg := fmt.Sprintf("error parsing authentication token: %s; access denied", err)
		log.Println(msg)
		return nil, fmt.Errorf(msg)
	}
	timeOk := (time.Now().Unix() - tokenTime) < (5 * 60)
	hostUUId := parts[0]
	if !timeOk {
		msg := fmt.Sprintf("authentication failed for Host UUID='%s': token has expired", hostUUId)
		log.Println(msg)
		return nil, fmt.Errorf(msg)
	}
	rows, err := r.db.Query("select * from pilotctl_is_admitted($1)", hostUUId)
	var hostname, hostIP string
	if err != nil {
		hostIP = parts[1]
		hostname = parts[2]
		msg := fmt.Sprintf("authentication failed for Host UUID='%s': cannot query admission table: %s\n"+
			"additional info: host IP = '%s', hostname = '%s'\n", hostUUId, err, hostIP, hostname)
		log.Println(msg)
		return nil, fmt.Errorf(msg)
	}
	var admitted bool
	for rows.Next() {
		err = rows.Scan(&admitted)
		if err != nil {
			msg := fmt.Sprintf("authentication failed for Host UUID='%s': %s\n", hostUUId, err)
			log.Println(msg)
			return nil, fmt.Errorf(msg)
		}
		break
	}

	// captures token information for remote host
	r.hostUUID = hostUUId
	r.hostname = hostname
	r.hostIP = hostIP

	if !admitted {
		// log an authentication error
		msg := fmt.Sprintf("authentication failed for Host UUID='%s', host has not been admitted to service\n", hostUUId)
		// no user principal is returned as authentication failed
		log.Println(msg)
		return nil, fmt.Errorf(msg)
	}

	// otherwise, returns a principal to signify that authentication succeeded
	return &h.UserPrincipal{
		// use a dummy email with the pilot host uuid as username
		Username: fmt.Sprintf("%s@pilot.com", hostUUId),
		// no access rights are required for pilot
		Rights:  h.Controls{},
		Created: time.Now(),
	}, nil
}

// AuthenticateUser authenticate user requests
func (r *API) AuthenticateUser(request http.Request) (*h.UserPrincipal, error) {
	// get the credentials from the request header
	user, pwd := h.ParseBasicToken(request)
	// validate the credentials and retrieve user access controls
	userPrincipal, err := r.iLink.Login(&ilink.Login{
		Username: user,
		Password: pwd,
	})
	if err != nil {
		msg := fmt.Sprintf("WARNING: user authentication failed, %s\n", err)
		log.Println(msg)
		return nil, fmt.Errorf(msg)
	}
	// return the user principal
	return userPrincipal, nil
}

// GetPackages get a list of packages in the backing Artisan registry
func (r *API) GetPackages() ([]PackageInfo, error) {
	// the URI to connect to the Artisan registry
	uri := fmt.Sprintf("%s/repository", r.conf.getArtRegUri())
	bytes, err := makeRequest(uri, "GET", r.conf.getArtRegUser(), r.conf.getArtRegPwd(), nil)
	if err != nil {
		return nil, err
	}
	var repos []registry.Repository
	err = json.Unmarshal(bytes, &repos)
	if err != nil {
		return nil, err
	}
	// removes protocol prefix from URI
	regDomain := r.conf.getArtRegUri()[strings.Index(r.conf.getArtRegUri(), "//")+2:]
	// removes forward slash at the end if it finds one
	if strings.HasSuffix(regDomain, "/") {
		regDomain = regDomain[0 : len(regDomain)-1]
	}
	result := make([]PackageInfo, 0)
	// gets the package filter is defined
	filter := r.conf.GetArtRegPackageFilter()
	for _, repo := range repos {
		// if a filter has been defined and the repository name does not start with the filter
		if len(filter) > 0 && !strings.HasPrefix(repo.Repository, filter) {
			// filter out from the result
			continue
		}
		pack := PackageInfo{
			Name: fmt.Sprintf("%s/%s", regDomain, repo.Repository),
			Tags: []TagInfo{},
		}
		for _, p := range repo.Packages {
			for _, tag := range p.Tags {
				pack.Tags = append(pack.Tags, TagInfo{
					Id:      p.Id,
					Name:    tag,
					Ref:     p.FileRef,
					Created: p.Created,
					Type:    p.Type,
					Size:    p.Size,
				})
			}
		}
		result = append(result, pack)
	}
	return result, nil
}

func (r *API) GetPackageAPI(name string) ([]*data.FxInfo, error) {
	n, err := core.ParseName(name)
	if err != nil {
		return nil, err
	}
	// the URI to connect to the Artisan registry
	uri := fmt.Sprintf("%s/package/manifest/%s/%s/%s", r.conf.getArtRegUri(), n.Group, n.Name, n.Tag)
	bytes, err := makeRequest(uri, "GET", r.conf.getArtRegUser(), r.conf.getArtRegPwd(), nil)
	if err != nil {
		return nil, err
	}
	var manif data.Manifest
	err = json.Unmarshal(bytes, &manif)
	if err != nil {
		return nil, err
	}
	if manif.Functions == nil {
		return make([]*data.FxInfo, 0), nil
	}
	return manif.Functions, nil
}

// PutCommand put the command in the Onix database
func (r *API) PutCommand(cmd *Cmd) error {
	var meta map[string]interface{}
	m := make(map[string]interface{}, 0)
	m["input"] = cmd.Input
	inputBytes, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("cannot marshal command input: %s", err)
	}
	err = json.Unmarshal(inputBytes, &meta)
	if err != nil {
		return fmt.Errorf("cannot unmarshal input bytes: %s", err)
	}
	result, err := r.iLink.PutItem(&ilink.Item{
		Key:         strings.Replace(cmd.Key, " ", "", -1),
		Name:        cmd.Key,
		Description: cmd.Description,
		Type:        "ART_FX",
		Meta:        meta,
		Attribute: map[string]interface{}{
			"PACKAGE": cmd.Package,
			"FX":      cmd.Function,
			// ensures credentials are for the registry tied to pilotctl
			"USER":          r.conf.getArtRegUser(),
			"PWD":           r.conf.getArtRegPwd(),
			"VERBOSE":       cmd.Verbose,
			"CONTAINERISED": cmd.Containerised,
		},
	})
	if result != nil && result.Error {
		return fmt.Errorf("cannot set command in Onix: %s\n", result.Message)
	}
	if err != nil {
		return fmt.Errorf("cannot set command in Onix: %s\n", err)
	}
	return nil
}

func (r *API) GetAllCommands() ([]Cmd, error) {
	items, err := r.iLink.GetItemsByType("ART_FX")
	if err != nil {
		return nil, fmt.Errorf("cannot get commands from Onix: %s", err)
	}
	cmds := make([]Cmd, 0)
	for _, item := range items.Values {
		input, err := getInputFromMap(item.Meta)
		if err != nil {
			return nil, fmt.Errorf("cannot transform input map: %s", err)
		}
		cmds = append(cmds, Cmd{
			Key:           item.Key,
			Description:   item.Description,
			Package:       item.GetStringAttr("PACKAGE"),
			Function:      item.GetStringAttr("FX"),
			User:          item.GetStringAttr("USER"),
			Pwd:           item.GetStringAttr("PWD"),
			Verbose:       item.GetBoolAttr("VERBOSE"),
			Containerised: item.GetBoolAttr("CONTAINERISED"),
			Input:         input,
		})

	}
	return cmds, nil
}

func (r *API) GetCommand(cmdName string) (*Cmd, error) {
	item, err := r.iLink.GetItem(&ilink.Item{Key: cmdName})
	if err != nil {
		return nil, fmt.Errorf("cannot get command with key '%s' from Onix: %s", cmdName, err)
	}
	input, err := getInputFromMap(item.Meta)
	if err != nil {
		return nil, fmt.Errorf("cannot transform input map: %s", err)
	}
	return &Cmd{
		Key:           item.Key,
		Description:   item.Description,
		Package:       item.GetStringAttr("PACKAGE"),
		Function:      item.GetStringAttr("FX"),
		User:          item.GetStringAttr("USER"),
		Pwd:           item.GetStringAttr("PWD"),
		Verbose:       item.GetBoolAttr("VERBOSE"),
		Containerised: item.GetBoolAttr("CONTAINERISED"),
		Input:         input,
	}, nil
}

func (r *API) DeleteCommand(cmdName string) (string, error) {
	result, err := r.iLink.DeleteItem(&ilink.Item{Key: cmdName})
	if err != nil {
		return "", fmt.Errorf("cannot delete command with key '%s' from Onix: %s", cmdName, err)
	}
	return result.Operation, nil
}

func (r *API) SetDictionary(dictionary Dictionary) (string, error) {
	result, err := r.iLink.PutItem(&ilink.Item{
		Key:         DKey(dictionary.Key),
		Name:        dictionary.Name,
		Description: dictionary.Description,
		Type:        "U_DICTIONARY",
		Meta:        dictionary.Values,
		Tag:         toInterfaceSlice(dictionary.Tags),
	})
	if result != nil && result.Error {
		return result.Operation, fmt.Errorf("cannot set dictionary in Onix CMDB: %s\n", result.Message)
	}
	if err != nil {
		return "", fmt.Errorf("cannot set dictionary in Onix CMDB: %s\n", err)
	}
	return result.Operation, nil
}

func toInterfaceSlice(tags []string) []interface{} {
	var result = make([]interface{}, 0)
	for _, tag := range tags {
		result = append(result, tag)
	}
	return result
}

func (r *API) DeleteDictionary(key string) (string, error) {
	result, err := r.iLink.DeleteItem(&ilink.Item{Key: DKey(key)})
	if err != nil {
		return "", fmt.Errorf("cannot delete command with key '%s' from Onix CMDB: %s", DKey(key), err)
	}
	return result.Operation, nil
}

func (r *API) GetDictionary(key string) (*Dictionary, error) {
	item, err := r.iLink.GetItem(&ilink.Item{Key: DKey(key)})
	if err != nil {
		return nil, fmt.Errorf("cannot get dictionary with key '%s' from Onix CMDB: %s", DKey(key), err)
	}
	return dict(*item), nil
}

func (r *API) GetDictionaries(values bool) ([]*Dictionary, error) {
	items, err := r.iLink.GetItemsByType("U_DICTIONARY")
	if err != nil {
		return nil, fmt.Errorf("cannot get dictionaries from Onix CMDB: %s", err)
	}
	result := make([]*Dictionary, len(items.Values))
	for i, item := range items.Values {
		result[i] = dict(item)
		// if no values are required
		if !values {
			// then ensure they are not set in the result
			result[i].Values = nil
		}
	}
	return result, nil
}

func (r *API) CompleteJob(status *JobResult) error {
	logMsg := status.Log
	// if there was a failure, and we have an error message, add it to the log
	if !status.Success && len(status.Err) > 0 {
		logMsg = fmt.Sprintf("%s !!! ERROR: %s\n", logMsg, status.Err)
	}
	return r.db.RunCommand("select pilotctl_complete_job($1, $2, $3)", status.JobId, logMsg, !status.Success)
}

func (r *API) GetAreas(orgGroup string) ([]Area, error) {
	items, err := r.iLink.GetChildrenByType(&ilink.Item{Key: orgGroup}, "U_AREA")
	if err != nil {
		return nil, err
	}
	var areas []Area
	for _, item := range items.Values {
		areas = append(areas, Area{
			Key:         item.Key,
			Name:        item.Name,
			Description: item.Description,
		})
	}
	return areas, nil
}

func (r *API) GetOrgs(orgGroup string) ([]Org, error) {
	items, err := r.iLink.GetChildrenByType(&ilink.Item{Key: orgGroup}, "U_ORG")
	if err != nil {
		return nil, err
	}
	var orgs []Org
	for _, item := range items.Values {
		orgs = append(orgs, Org{
			Key:         item.Key,
			Name:        item.Name,
			Description: item.Description,
		})
	}
	return orgs, nil
}

func (r *API) GetLocations(area string) ([]Location, error) {
	items, err := r.iLink.GetChildrenByType(&ilink.Item{Key: area}, "U_LOCATION")
	if err != nil {
		return nil, err
	}
	var orgs []Location
	for _, item := range items.Values {
		orgs = append(orgs, Location{
			Key:  item.Key,
			Name: item.Name,
		})
	}
	return orgs, nil
}

func (r *API) GetOrgGroups() ([]Org, error) {
	items, err := r.iLink.GetItemsByType("U_ORG_GROUP")
	if err != nil {
		return nil, err
	}
	var orgs []Org
	for _, item := range items.Values {
		orgs = append(orgs, Org{
			Key:         item.Key,
			Name:        item.Name,
			Description: item.Description,
		})
	}
	return orgs, nil
}

func (r *API) CreateJobBatch(info JobBatchInfo) (int64, error) {
	if len(info.HostUUID) == 0 {
		return -1, fmt.Errorf("host UUID is missing\n")
	}
	if len(info.FxKey) == 0 {
		return -1, fmt.Errorf("fx is missing\n")
	}
	// create a job batch identifier
	rows, err := r.db.Query("select * from pilotctl_create_job_batch($1, $2, $3, $4)", info.Name, info.Notes, "???", info.Label)
	if err != nil {
		return -1, fmt.Errorf("cannot create job batch: %s\n", err)
	}
	var batchId int64 = -1
	for rows.Next() {
		rows.Scan(&batchId)
	}
	if batchId == -1 {
		return -1, fmt.Errorf("cannot retrieve job batch Id\n")
	}
	// add jobs to the batch using the batch ID
	var returnError error
	for _, uuid := range info.HostUUID {
		err = r.db.RunCommand("select pilotctl_create_job($1, $2, $3, $4)", batchId, uuid, info.FxKey, info.FxVersion)
		// if there is an error creating the job
		if err != nil {
			if returnError == nil {
				returnError = err
			}
			// accumulates the error and continue with the next job
			returnError = fmt.Errorf("can't create job: %s\n", returnError)
		}
	}
	// return any job creation error
	return batchId, returnError
}

func (r *API) GetJobs(oGroup, or, ar, loc string, batchId *int64) ([]Job, error) {
	jobs := make([]Job, 0)
	rows, err := r.db.Query("select * from pilotctl_get_jobs($1, $2, $3, $4, $5)", oGroup, or, ar, loc, batchId)
	if err != nil {
		return nil, fmt.Errorf("cannot get jobs: %s\n", err)
	}
	var (
		id         int64
		hostUUID   string
		jobBatchId int64
		fxKey      string
		fxVersion  int64
		created    sql.NullTime
		started    sql.NullTime
		completed  sql.NullTime
		log        sql.NullString
		e          sql.NullBool
		orgGroup   sql.NullString
		org        sql.NullString
		area       sql.NullString
		location   sql.NullString
		tag        []string
	)
	for rows.Next() {
		err = rows.Scan(&id, &hostUUID, &jobBatchId, &fxKey, &fxVersion, &created, &started, &completed, &log, &e, &orgGroup, &org, &area, &location, &tag)
		if err != nil {
			return nil, fmt.Errorf("cannot scan job row: %e\n", err)
		}
		jobs = append(jobs, Job{
			Id:         id,
			HostUUID:   hostUUID,
			JobBatchId: jobBatchId,
			FxKey:      fxKey,
			FxVersion:  fxVersion,
			Created:    timeF(created),
			Started:    timeF(started),
			Completed:  timeF(completed),
			Log:        stringF(log),
			Error:      boolF(e),
			OrgGroup:   orgGroup.String,
			Org:        org.String,
			Area:       area.String,
			Location:   location.String,
			Tag:        tag,
		})
	}
	return jobs, rows.Err()
}

func (r *API) GetHost(uuid string) (*Host, error) {
	rows, err := r.db.Query("select * from pilotctl_get_host($1)", uuid)
	if err != nil {
		return nil, fmt.Errorf("cannot get host data: %s\n", err)
	}
	var (
		orgGroup string
		org      string
		area     string
		location string
		labels   []string
	)
	for rows.Next() {
		err = rows.Scan(&orgGroup, &org, &area, &location, &labels)
		if err != nil {
			return nil, fmt.Errorf("cannot scan host row: %e\n", err)
		}
		return &Host{HostUUID: uuid, OrgGroup: orgGroup, Org: org, Area: area, Location: location, Label: labels}, nil
	}
	return nil, fmt.Errorf("host uuid '%s' cannot be found in data source\n", uuid)
}

func (r *API) GetJobBatches(name, owner *string, from, to *time.Time, label *[]string) ([]JobBatch, error) {
	batches := make([]JobBatch, 0)
	rows, err := r.db.Query("select * from pilotctl_get_job_batches($1, $2, $3, $4, $5)", name, from, to, label, owner)
	if err != nil {
		return nil, fmt.Errorf("cannot get job batches: %s\n", err)
	}
	var (
		id      int64
		name2   string
		notes   string
		created sql.NullTime
		owner2  string
		labels  []string
		jobs    int
	)
	for rows.Next() {
		err = rows.Scan(&id, &name2, &notes, &labels, &created, &owner2, &jobs)
		if err != nil {
			return nil, fmt.Errorf("cannot scan job batch row: %e\n", err)
		}
		batches = append(batches, JobBatch{
			BatchId: id,
			Name:    name2,
			Notes:   notes,
			Label:   labels,
			Owner:   owner2,
			Jobs:    jobs,
			Created: created.Time,
		})
	}
	return batches, rows.Err()
}

func (r *API) Augment(events *Events) (*Events, error) {
	// as all events come from the same host get the host UUID from the first event
	hostUUID := events.Events[0].HostUUID
	// retrieve host information
	host, err := r.GetHost(hostUUID)
	if err != nil {
		return events, err
	}
	// add info to events
	var result []Event
	for _, event := range events.Events {
		ev := Event{
			EventID:           event.EventID,
			Client:            event.Client,
			Hostname:          event.Hostname,
			HostUUID:          event.HostUUID,
			MachineId:         event.MachineId,
			HostAddress:       event.HostAddress,
			Organisation:      host.Org,
			OrganisationGroup: host.OrgGroup,
			Area:              host.Area,
			Location:          host.Location,
			Facility:          event.Facility,
			Priority:          event.Priority,
			Severity:          event.Severity,
			Time:              event.Time,
			TLSPeer:           event.TLSPeer,
			BootTime:          event.BootTime,
			Content:           event.Content,
			Tag:               event.Tag,
			MacAddress:        event.MacAddress,
			HostLabel:         host.Label,
		}
		result = append(result, ev)
	}
	return &Events{Events: result}, nil
}

// Login if the user is authenticated returns a list of access controls otherwise an error is returned
func (r *API) Login(username string) (h.Controls, error) {
	user, err := r.iLink.GetUser(&ilink.User{Key: username})
	if err != nil {
		return nil, fmt.Errorf("login failed for user '%s': %s\n", username, err)
	}
	// filter the controls using the realm
	var (
		controls = user.Controls()
		result   []h.Control
	)
	for _, control := range controls {
		if control.Realm == "pilotcl" {
			result = append(result, h.Control{
				Realm:  control.Realm,
				URI:    control.URI,
				Method: []string{control.Method},
			})
		}
	}
	return result, nil
}

func (r *API) AuthenticateActivationSvc(request http.Request) (*h.UserPrincipal, error) {
	cf := &Conf{}
	user, pwd := h.ParseBasicToken(request)
	if user == cf.GetActivationUser() && pwd == cf.GetActivationPwd() {
		return &h.UserPrincipal{
			Username: user,
			Rights:   nil,
			Created:  time.Now(),
			Context:  nil,
		}, nil
	}
	return nil, fmt.Errorf("invalid credentials")
}

func (r *API) UndoRegistration(mac string) error {
	if len(mac) == 0 {
		return fmt.Errorf("MAC-ADDRESS is missing")
	}
	return r.db.RunCommand("select pilotctl_unset_registration($1)", mac)
}

func (r *API) DecommissionHost(hostUUID string) error {
	if len(hostUUID) == 0 {
		return fmt.Errorf("HOST-UUID is missing")
	}
	// set a decom date for the host
	err := r.db.RunCommand("select pilotctl_decom_host($1)", hostUUID)
	if err != nil {
		return err
	}
	// delete host from cmdb
	_, err = r.iLink.DeleteItem(&ilink.Item{Key: strings.ToUpper(fmt.Sprintf("HOST:%s", hostUUID))})
	return err
}

func (r *API) UpsertCVE(hostUUID string, rep *CveReport) error {
	scanDate := time.Now().UTC()
	for _, cve := range rep.Cves {
		err := r.db.RunCommand("select pilotctl_unlink_cve($1, $2)", hostUUID, cve.Id)
		if err != nil {
			return fmt.Errorf("cannot unlink cve %s from host %s: %s", cve.Id, hostUUID, err)
		}
		err = r.db.RunCommand("select pilotctl_set_cve($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)",
			cve.Id,
			cve.Summary,
			cve.Fixed(),
			cve.CVSSScore,
			cve.CVSSType,
			cve.CVSSSeverity,
			cve.CVSSVector,
			cve.PrimarySrc,
			cve.Mitigations,
			cve.PatchURLs,
			cve.Confidence,
			cve.CPE,
			cve.References)
		if err != nil {
			return fmt.Errorf("cannot update cve %s: %s", cve.Id, err)
		}
		for _, affectedPackage := range cve.AffectedPackages {
			err = r.db.RunCommand("select pilotctl_set_cve_package($1, $2, $3, $4)",
				cve.Id,
				affectedPackage.Name,
				!affectedPackage.NotFixedYet,
				affectedPackage.FixedIn,
			)
			if err != nil {
				return fmt.Errorf("cannot update package %s for cve %s: %s", affectedPackage.Name, cve.Id, err)
			}
		}
		err = r.db.RunCommand("select pilotctl_link_cve($1, $2, $3)",
			hostUUID,
			cve.Id,
			scanDate,
		)
		if err != nil {
			return fmt.Errorf("cannot link cve %s to host %s: %s", cve.Id, hostUUID, err)
		}
	}
	err := r.db.RunCommand("select pilotctl_set_host_cve($1, $2, $3, $4, $5)",
		hostUUID,
		rep.Critical(),
		rep.High(),
		rep.Medium(),
		rep.Low(),
	)
	if err != nil {
		return fmt.Errorf("cannot update cve stats on host %s: %s", hostUUID, err)
	}
	return nil
}

func (r *API) GetCVEBaseline(score float64, label []string) ([]CvePackage, error) {
	rows, err := r.db.Query("select * from pilotctl_get_cve_baseline($1, $2)", score, label)
	if err != nil {
		return nil, fmt.Errorf("cannot get CVE baseline: %s\n", err)
	}
	var (
		hostUUID, cveID, packageName, fixedIn string
		cvssScore                             float64
	)
	var list []CvePackage
	for rows.Next() {
		err = rows.Scan(&hostUUID, &cveID, &packageName, &fixedIn, &cvssScore)
		if err != nil {
			return nil, fmt.Errorf("cannot scan CVE baseline row: %e\n", err)
		}
		list = append(list, CvePackage{
			HostUUID:    hostUUID,
			CveID:       cveID,
			PackageName: packageName,
			FixedIn:     fixedIn,
			CvssScore:   cvssScore,
		})
	}
	return list, nil
}

// SubmitMetrics post the metrics to a backend using a connector
func (r *API) SubmitMetrics(channel string, content []byte) ConnResult {
	var (
		pbUnmarshall = pmetric.ProtoUnmarshaler{}
		conn         string
		ok           bool
		dpConv       = NewOtelDataPointConverter()
	)
	// find the connector to use based on the specified channel
	if conn, ok = connectorName(channel); !ok {
		return ConnResult{
			Error:             fmt.Sprintf("telemetry connector not defined, skipping telemetry recording"),
			TotalEntries:      -1,
			SuccessfulEntries: -1,
		}
	}
	files, err := os2.ReadFileBatchFromBytes(content)
	if err != nil {
		return ConnResult{
			Error:             fmt.Sprintf("cannot debatch telemetry metrics: %s", err),
			TotalEntries:      -1,
			SuccessfulEntries: -1,
		}
	}
	result := new(ConnResult)
	for i, file := range files {
		// unmarshall the protobuf content
		metrics, err := pbUnmarshall.UnmarshalMetrics(file)
		if err != nil {
			return ConnResult{
				Error:             fmt.Sprintf("cannot unmarshal protobuf metrics, corrupted or invalid format: %s", err),
				TotalEntries:      len(files),
				SuccessfulEntries: i,
			}
		}
		// converts content to json
		dataPoints, err := dpConv.Convert(metrics)
		if err != nil {
			return ConnResult{
				Error:             fmt.Sprintf("cannot convert open telemetry metrics to data points: %s", err),
				TotalEntries:      len(files),
				SuccessfulEntries: i,
			}
		}
		data, err := json.Marshal(dataPoints)
		if err != nil {
			return ConnResult{
				Error:             fmt.Sprintf("cannot marshal data points to json: %s", err),
				TotalEntries:      len(files),
				SuccessfulEntries: i,
			}
		}
		pointsFile := dataPointsFilename()
		err = os.WriteFile(pointsFile, data, 0755)
		if err != nil {
			return ConnResult{
				Error:             fmt.Sprintf("cannot write data points to json file: %s", err),
				TotalEntries:      len(files),
				SuccessfulEntries: i,
			}
		}
		// execute the connector passing in the json content
		dir, _ := os.Getwd()
		c := exec.Command(fmt.Sprintf("%s/%s", dir, conn), pointsFile)
		c.Env = os.Environ()
		outBytes, err := c.Output()
		if err != nil {
			return ConnResult{
				Error:             fmt.Sprintf("connector %s error: %s", conn, err),
				TotalEntries:      len(files),
				SuccessfulEntries: i,
			}
		}
		_ = os.Remove(pointsFile)
		// extracts the response for the stdout
		out := regexp.MustCompile("{\"e\":.*}").FindString(string(outBytes[:]))
		if len(out) == 0 {
			return ConnResult{
				Error:             fmt.Sprintf("connector %s gave an empty response, execution failed: %s", conn, err),
				TotalEntries:      len(files),
				SuccessfulEntries: i,
			}
		}
		err = json.Unmarshal([]byte(out), result)
		if err != nil {
			return ConnResult{
				Error:             fmt.Sprintf("connector %s gave an invalid result format, unmarshalling failed: %s", conn, err),
				TotalEntries:      len(files),
				SuccessfulEntries: i,
			}
		}
		if len(result.Error) > 0 {
			return ConnResult{
				Error:             fmt.Sprintf("connector %s returned error: %s", conn, result.Error),
				TotalEntries:      len(files),
				SuccessfulEntries: i,
			}
		}
	}
	return *result
}

// dataPointsFilename works out a unique filename for the datapoints file in the telemetry buffer folder
func dataPointsFilename() string {
	cfg := NewConf()
	path, _ := filepath.Abs(cfg.getTelemBufferPath())
	return filepath.Join(path, fmt.Sprintf("%d.json", time.Now().UTC().UnixNano()))
}

// SubmitLogs post the logs to a backend using a connector
// TODO: use file buffer to pass logs to connector as done in SubmitMetrics
func (r *API) SubmitLogs(channel string, content []byte) ConnResult {
	var (
		pbUnmarshall = plog.ProtoUnmarshaler{}
		jsonMarshal  = plog.JSONMarshaler{}
		conn         string
		ok           bool
	)
	// find the connector to use based on the specified channel
	if conn, ok = connectorName(channel); !ok {
		return ConnResult{
			Error:             fmt.Sprintf("telemetry connector not defined, skipping telemetry recording"),
			TotalEntries:      -1,
			SuccessfulEntries: -1,
		}
	}
	files, err := os2.ReadFileBatchFromBytes(content)
	if err != nil {
		return ConnResult{
			Error:             fmt.Sprintf("cannot debatch telemetry logs: %s", err),
			TotalEntries:      -1,
			SuccessfulEntries: -1,
		}
	}
	result := new(ConnResult)
	for i, file := range files {
		// unmarshall the protobuf content
		logs, err := pbUnmarshall.UnmarshalLogs(file)
		if err != nil {
			return ConnResult{
				Error:             fmt.Sprintf("cannot unmarshal logs: %s", err),
				TotalEntries:      len(files),
				SuccessfulEntries: i,
			}
		}
		// converts content to json
		data, err := jsonMarshal.MarshalLogs(logs)
		if err != nil {
			return ConnResult{
				Error:             fmt.Sprintf("cannot marshal request to json: %s", err),
				TotalEntries:      len(files),
				SuccessfulEntries: i,
			}
		}
		// base 64 encode the json content
		b64Data := base64.StdEncoding.EncodeToString(data)
		// execute the connector passing in the json content
		dir, _ := os.Getwd()
		c := exec.Command(fmt.Sprintf("%s/%s", dir, conn), b64Data)
		c.Env = os.Environ()
		outBytes, err := c.Output()
		out := regexp.MustCompile("{\"e\":.*}").FindString(string(outBytes[:]))
		if len(out) == 0 {
			return ConnResult{
				Error:             fmt.Sprintf("connector %s gave an empty response, execution failed: %s", conn, err),
				TotalEntries:      len(files),
				SuccessfulEntries: i,
			}
		}
		err = json.Unmarshal([]byte(out), result)
		if err != nil {
			return ConnResult{
				Error:             fmt.Sprintf("connector %s gave an invalid result format, unmarshalling failed: %s", conn, err),
				TotalEntries:      len(files),
				SuccessfulEntries: i,
			}
		}
		if len(result.Error) > 0 {
			return ConnResult{
				Error:             fmt.Sprintf("connector %s returned error: %s", conn, result.Error),
				TotalEntries:      len(files),
				SuccessfulEntries: i,
			}
		}
	}
	return *result
}

func reverse(str string) (result string) {
	for _, v := range str {
		result = string(v) + result
	}
	return
}

func timeF(t sql.NullTime) string {
	if t.Valid {
		return t.Time.Format(time.RFC822Z)
	}
	return ""
}

func stringF(t sql.NullString) string {
	if t.Valid {
		return t.String
	}
	return ""
}

func boolF(t sql.NullBool) bool {
	if t.Valid {
		return t.Bool
	}
	return false
}

func DKey(key string) string {
	return fmt.Sprintf("DC:%s", strings.ToUpper(key))
}

func dict(item ilink.Item) *Dictionary {
	return &Dictionary{
		Key:         item.Key[3:],
		Name:        item.Name,
		Description: item.Description,
		Values:      item.Meta,
		Tags:        fromInterfaceSlice(item.Tag),
	}
}

func fromInterfaceSlice(tags []interface{}) []string {
	var result = make([]string, len(tags))
	for i, tag := range tags {
		result[i] = fmt.Sprint(tag)
	}
	return result
}

// the name of the connector to use based on the specified channel
func connectorName(channel string) (string, bool) {
	// the value of the env is of the format channel:connector,... (e.g. ch1:mongodb_conn,ch2:pravega_conn)
	cfg := NewConf()
	n := cfg.get(ConfTelemConnectors)
	if len(n) == 0 {
		return "", false
	}
	parts := strings.Split(n, ",")
	for _, part := range parts {
		subParts := strings.Split(part, ":")
		if strings.EqualFold(subParts[0], channel) {
			return subParts[1], true
		}
	}
	return "", false
}

// ConnResult the result of the connector execution
type ConnResult struct {
	Error             string `json:"e,omitempty"`
	TotalEntries      int    `json:"t,omitempty"`
	SuccessfulEntries int    `json:"s,omitempty"`
}
