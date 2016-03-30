package main

import (
	"errors"
	"flag"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

//example of complex Struct
type ownerInfo struct {
	Name         string    `long:"name" description:"overwrite owner name"`
	Organization string    `long:"org" description:"overwrite owner organisation"`
	Bio          string    `long:"bio" description:"overwrite owner biography"`
	Dob          time.Time `long:"dob" description:"overwrite owner date of birth"`
}
type databaseInfo struct {
	Server        string `long:"srv" description:"overwrite database server ip address"`
	ConnectionMax int    `long:"comax" description:"overwrite maximum number of connection on the database"`
	Enable        bool   `long:"ena" description:"overwrite database enable"`
}
type serverInfo struct {
	IP string `long:"ip" description:"overwrite server ip address"`
	Dc string `long:"dc" description:"overwrite server domain controller"`
}
type clientInfo struct {
	Data  []int        `long:"data" description:"overwrite clients data"`
	Hosts []serverInfo `description:"overwrite clients host names"`
}
type example struct {
	Title    string       `short:"t" description:"overwrite title"` //
	Owner    ownerInfo    `long:"own"  description:"overwrite server ip address"`
	Database databaseInfo ` description:"overwrite server ip address"`
	Servers  serverInfo   `description:"overwrite servers info --servers.[ip|dc] [srv name]: value"`
	Clients  *clientInfo  `long:"cli"  description:"overwrite server ip address"`
}

func TestGetTypesRecursive(t *testing.T) {
	//Test all
	var ex1 example
	namesmap := make(map[string]reflect.Type)
	if err := getTypesRecursive(reflect.ValueOf(&ex1), namesmap, ""); err != nil {
		t.Errorf("Error %s", err.Error())
	}

	checkType := map[string]reflect.Type{
		"title":          reflect.TypeOf(""),
		"own":            reflect.TypeOf(ownerInfo{}),
		"cli":            reflect.TypeOf(&clientInfo{}),
		"cli.hosts.ip":   reflect.TypeOf(""),
		"t":              reflect.TypeOf(""),
		"database":       reflect.TypeOf(databaseInfo{}),
		"cli.data":       reflect.TypeOf([]int{}),
		"cli.hosts":      reflect.TypeOf([]serverInfo{}),
		"cli.hosts.dc":   reflect.TypeOf(""),
		"own.name":       reflect.TypeOf(""),
		"own.bio":        reflect.TypeOf(""),
		"own.dob":        reflect.TypeOf(time.Time{}),
		"database.srv":   reflect.TypeOf(""),
		"database.comax": reflect.TypeOf(0),
		"servers":        reflect.TypeOf(serverInfo{}),
		"own.org":        reflect.TypeOf(""),
		"database.ena":   reflect.TypeOf(true),
		"servers.ip":     reflect.TypeOf(""),
		"servers.dc":     reflect.TypeOf(""),
	}
	for name, nameType := range namesmap {
		if checkType[name] != nameType {
			t.Fatalf("Tag : %s, got %s expected %s\n", name, nameType, checkType[name])
		}
	}

}

func TestParseArgs(t *testing.T) {
	//creating parsers
	parsers := map[reflect.Type]flag.Getter{}
	var myStringParser stringValue
	var myBoolParser boolValue
	var myIntParser intValue
	var myCustomParser customValue
	var mySliceServerParser sliceServerValue
	var myTimeParser timeValue
	parsers[reflect.TypeOf("")] = &myStringParser
	parsers[reflect.TypeOf(true)] = &myBoolParser
	parsers[reflect.TypeOf(1)] = &myIntParser
	parsers[reflect.TypeOf([]int{})] = &myCustomParser
	parsers[reflect.TypeOf([]serverInfo{})] = &mySliceServerParser
	parsers[reflect.TypeOf(time.Now())] = &myTimeParser

	//Test all
	var ex1 example
	tagsmap := make(map[string]reflect.Type)

	if err := getTypesRecursive(reflect.ValueOf(ex1), tagsmap, ""); err != nil {
		t.Errorf("Error %s", err.Error())
	}

	args := []string{
		// "-title", "myTitle",
		// "own",
		// "cli":
		"-cli.hosts", "{myIp1,myDc1}",
		"-t", "myTitle",
		// "-database",""
		"-cli.data", "{1,2,3,4}",
		// "-cli.hosts",""
		"-cli.hosts", "{myIp2,myDc2}",
		"-own.name", "myOwnName",
		"-own.bio", "myOwnBio",
		"-own.dob", "1979-05-27T07:32:00Z",
		"-database.srv", "mySrv",
		"-database.comax", "1000",
		// "-servers":
		"-own.org", "myOwnOrg",
		"-database.ena", //=true"
		"-servers.ip", "myServersIp",
		"-servers.dc", "myServersDc",
	}
	pargs, err := parseArgs(args, tagsmap, parsers)
	if err != nil {
		t.Errorf("Error %s", err.Error())
	}
	// fmt.Printf("result:%+v\n", pargs)

	//CHECK

	cliHostsCheck := sliceServerValue([]serverInfo{{"myIp1", "myDc1"}, {"myIp2", "myDc2"}})
	tCheck := stringValue("myTitle")
	cliDataCheck := customValue([]int{1, 2, 3, 4})
	ownNameCheck := stringValue("myOwnName")
	ownBioCheck := stringValue("myOwnBio")
	dob, _ := time.Parse(time.RFC3339, "1979-05-27T07:32:00Z")
	ownDobCheck := timeValue(dob)
	databaseSrvCheck := stringValue("mySrv")
	databaseComaxCheck := intValue(1000)
	ownOrgCheck := stringValue("myOwnOrg")
	databaseEnaCheck := boolValue(true)
	serversIPCheck := stringValue("myServersIp")
	serversDcCheck := stringValue("myServersDc")

	checkParse := map[string]flag.Getter{

		// "title", "myTitle",
		// "own",
		// "cli":
		"cli.hosts": &cliHostsCheck,
		"t":         &tCheck,
		// "database",""
		"cli.data": &cliDataCheck,
		// "cli.hosts",""
		"own.name":       &ownNameCheck,
		"own.bio":        &ownBioCheck,
		"own.dob":        &ownDobCheck,
		"database.srv":   &databaseSrvCheck,
		"database.comax": &databaseComaxCheck,
		// "servers":
		"own.org":      &ownOrgCheck,
		"database.ena": &databaseEnaCheck, //=true"
		"servers.ip":   &serversIPCheck,
		"servers.dc":   &serversDcCheck,
	}

	for tag, inter := range pargs {

		if !reflect.DeepEqual(checkParse[tag].Get(), inter.Get()) {
			t.Fatalf("Error tag %s : expected %+v got %+v", tag, checkParse[tag].Get(), inter.Get())
		}
	}
}

func TestFillStructRecursive(t *testing.T) {
	//creating parsers
	parsers := map[reflect.Type]flag.Getter{}
	var myStringParser stringValue
	var myBoolParser boolValue
	var myIntParser intValue
	var myCustomParser sliceIntValue
	var mySliceServerParser sliceServerValue
	var myTimeParser timeValue
	parsers[reflect.TypeOf("")] = &myStringParser
	parsers[reflect.TypeOf(true)] = &myBoolParser
	parsers[reflect.TypeOf(1)] = &myIntParser
	parsers[reflect.TypeOf([]int{})] = &myCustomParser
	parsers[reflect.TypeOf([]serverInfo{})] = &mySliceServerParser
	parsers[reflect.TypeOf(time.Now())] = &myTimeParser

	//Test all
	var ex1 example
	tagsmap := make(map[string]reflect.Type)

	if err := getTypesRecursive(reflect.ValueOf(&ex1), tagsmap, ""); err != nil {
		t.Errorf("Error %s", err.Error())
	}
	args := []string{
		// "-title", "myTitle",
		// "own",
		// "cli":
		"-cli.hosts", "{myIp1,myDc1}",
		"-t", "myTitle",
		// "-database",""
		// "-cli.hosts",""
		"-cli.hosts", "{myIp2,myDc2}",
		"-own.name", "myOwnName",
		"-own.bio", "myOwnBio",
		"-own.dob", "1979-05-27T07:32:00Z",
		"-database.srv", "mySrv",
		"-database.comax", "1000",
		// "-servers":
		"-own.org", "myOwnOrg",
		"-database.ena", //=true"
		"-servers.ip", "myServersIp",
		"-servers.dc", "myServersDc",
		"-cli.data", "1",
		"-cli.data", "2",
		"-cli.data", "3",
		"-cli.data", "4",
	}

	pargs, err := parseArgs(args, tagsmap, parsers)
	if err != nil {
		t.Errorf("Error %s", err.Error())
	}

	if err := fillStructRecursive(reflect.ValueOf(&ex1), pargs, ""); err != nil {
		t.Errorf("Error %s", err.Error())
	}

	//CHECK
	var check example
	check.Title = "myTitle"
	check.Owner.Name = "myOwnName"
	check.Owner.Organization = "myOwnOrg"
	check.Owner.Bio = "myOwnBio"
	check.Owner.Dob, _ = time.Parse(time.RFC3339, "1979-05-27T07:32:00Z")
	check.Database.Server = "mySrv"
	check.Database.ConnectionMax = 1000
	check.Database.Enable = true
	check.Servers.IP = "myServersIp"
	check.Servers.Dc = "myServersDc"
	check.Clients = &clientInfo{Data: []int{1, 2, 3, 4}, Hosts: []serverInfo{{"myIp1", "myDc1"}, {"myIp2", "myDc2"}}}
	if !reflect.DeepEqual(ex1, check) {
		t.Fatalf("\nexpected\t: %+v\ngot\t\t\t: %+v", check, ex1)
	}

}

// -- CUSTOM PARSERS
// -- custom Value
type customValue []int

func bracket(r rune) bool {
	return r == '{' || r == '}' || r == ',' || r == ';'
}
func (c *customValue) Set(s string) error {
	tabStr := strings.FieldsFunc(s, bracket)
	for _, str := range tabStr {
		v, err := strconv.Atoi(str)
		if err != nil {
			return err
		}
		*c = append(*c, v)
	}
	return nil
}

func (c *customValue) Get() interface{} { return []int(*c) }

func (c *customValue) String() string { return fmt.Sprintf("%v", *c) }

// -- sliceIntValue
type sliceIntValue []int

func (c *sliceIntValue) Set(s string) error {
	v, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	*c = append(*c, v)
	return nil
}

func (c *sliceIntValue) Get() interface{} { return []int(*c) }

func (c *sliceIntValue) String() string { return fmt.Sprintf("%v", *c) }

// -- sliceServerValue format {IP,DC}
type sliceServerValue []serverInfo

func (c *sliceServerValue) Set(s string) error {
	tabStr := strings.FieldsFunc(s, bracket)
	if len(tabStr) != 2 {
		return errors.New("sliceServerValue cannot parse %s to serverInfo. Format {IP,DC}")
	}
	srv := serverInfo{IP: tabStr[0], Dc: tabStr[1]}
	*c = append(*c, srv)
	return nil
}

func (c *sliceServerValue) Get() interface{} { return []serverInfo(*c) }

func (c *sliceServerValue) String() string { return fmt.Sprintf("%v", *c) }

func TestLoadParsers(t *testing.T) {
	//creating parsers
	customParsers := map[reflect.Type]flag.Getter{}
	var mySliceIntParser sliceIntValue
	var mySliceServerParser sliceServerValue
	customParsers[reflect.TypeOf([]int{})] = &mySliceIntParser
	customParsers[reflect.TypeOf([]serverInfo{})] = &mySliceServerParser
	//Test loadParsers
	parsers, err := loadParsers(customParsers)
	if err != nil {
		t.Errorf("Error %s", err.Error())
	}

	//check
	check := map[reflect.Type]flag.Getter{}
	check[reflect.TypeOf([]int{})] = &mySliceIntParser
	check[reflect.TypeOf([]serverInfo{})] = &mySliceServerParser
	var stringParser stringValue
	var boolParser boolValue
	var intParser intValue
	var timeParser timeValue
	check[reflect.TypeOf("")] = &stringParser
	check[reflect.TypeOf(true)] = &boolParser
	check[reflect.TypeOf(1)] = &intParser
	check[reflect.TypeOf(time.Now())] = &timeParser

	if !reflect.DeepEqual(parsers, check) {
		t.Fatalf("\nexpected\t: %+v\ngot\t\t\t: %+v", check, parsers)
	}

}

func TestLoad(t *testing.T) {
	//creating parsers
	customParsers := map[reflect.Type]flag.Getter{}
	var mySliceIntParser sliceIntValue
	var mySliceServerParser sliceServerValue
	customParsers[reflect.TypeOf([]int{})] = &mySliceIntParser
	customParsers[reflect.TypeOf([]serverInfo{})] = &mySliceServerParser

	//args
	args := []string{
		"-cli.hosts", "{myIp1,myDc1}",
		"-t", "myTitle",
		"-cli.hosts", "{myIp2,myDc2}",
		"-own.name", "myOwnName",
		"-own.bio", "myOwnBio",
		"-own.dob", "1979-05-27T07:32:00Z",
		"-database.srv", "mySrv",
		"-database.comax", "1000",
		"-own.org", "myOwnOrg",
		"-database.ena", //=true"
		"-servers.ip", "myServersIp",
		"-servers.dc", "myServersDc",
		"-cli.data", "1",
		"-cli.data", "2",
		"-cli.data", "3",
		"-cli.data", "4",
	}

	//Test all
	var ex1 example

	if err := Load(&ex1, args, customParsers); err != nil {
		t.Errorf("Error %s", err.Error())
	}
	//CHECK
	var check example
	check.Title = "myTitle"
	check.Owner.Name = "myOwnName"
	check.Owner.Organization = "myOwnOrg"
	check.Owner.Bio = "myOwnBio"
	check.Owner.Dob, _ = time.Parse(time.RFC3339, "1979-05-27T07:32:00Z")
	check.Database.Server = "mySrv"
	check.Database.ConnectionMax = 1000
	check.Database.Enable = true
	check.Servers.IP = "myServersIp"
	check.Servers.Dc = "myServersDc"
	check.Clients = &clientInfo{Data: []int{1, 2, 3, 4}, Hosts: []serverInfo{{"myIp1", "myDc1"}, {"myIp2", "myDc2"}}}
	if !reflect.DeepEqual(ex1, check) {
		t.Fatalf("\nexpected\t: %+v\ngot\t\t\t: %+v", check, ex1)
	}
}
