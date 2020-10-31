// Copyright 2019 Adam Wasila
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package arangodbadapter

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/arangodb/go-driver"
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	. "github.com/smartystreets/goconvey/convey"
)

func ExampleNewAdapter() {
	a, err := NewAdapter(
		OpCollectionName("casbinrules_example"),
		OpFieldMapping("p", "sub", "obj", "act"))

	if err != nil {
		fmt.Printf("Adapter creation error! %s\n", err)
		return
	}

	m, err := model.NewModelFromString(`
	[request_definition]
	r = sub, obj, act
	
	[policy_definition]
	p = sub, obj, act
	
	[policy_effect]
	e = some(where (p.eft == allow))
	
	[matchers]
	m = r.sub == p.sub && r.obj == p.obj && r.act == p.act
	`)
	if err != nil {
		fmt.Printf("Enforcer creation error! %s\n", err)
		return
	}

	e, err := casbin.NewEnforcer(m, a)
	if err != nil {
		fmt.Printf("Enforcer creation error! %s\n", err)
		return
	}
	err = e.LoadPolicy()
	if err != nil {
		fmt.Printf("Load policy error! %s\n", err)
		return
	}

	sub, obj, act := "adam", "data1", "read"

	_, _ = e.AddPolicy("adam", "data1", "read")
	_ = e.SavePolicy()

	r, err := e.Enforce(sub, obj, act)
	if err != nil {
		fmt.Printf("Failed to enforce! %s\n", err)
		return
	}
	if !r {
		fmt.Printf("%s %s %s: Forbidden!\n", sub, obj, act)
	} else {
		fmt.Printf("%s %s %s: Access granted\n", sub, obj, act)
	}
	// Output:
	// adam data1 read: Access granted
}

func TestArangodbNewAdapter(t *testing.T) {
	var operatorstests = []struct {
		name        string
		in          []adapterOption
		expectedErr func(error) bool
	}{
		{"Custom Endpoint", []adapterOption{OpEndpoints("http://localhost:8529")}, nil},
		{"Custom Database Name", []adapterOption{OpDatabaseName("casbin")}, nil},
		{"Custom Collection Name", []adapterOption{OpCollectionName("casbin_rules")}, nil},
		{"Custom Field Mapping", []adapterOption{OpFieldMapping("p", "sub", "obj", "act")}, nil},
		{"Autocreate", []adapterOption{OpAutocreate(false)}, nil},
		{"Basic Auth Credentials", []adapterOption{OpBasicAuthCredentials("root", "password")}, nil},
		{"Basic Auth Credentials - passing wrong credentials to database with auth", []adapterOption{
			OpEndpoints("http://localhost:8530"),
			OpBasicAuthCredentials("root", "wrongpassword"),
		}, func(err error) bool {
			return driver.IsUnauthorized(err)
		}},
		{"Basic Auth Credentials - passing good credentials to database with auth", []adapterOption{
			OpEndpoints("http://localhost:8530"),
			OpBasicAuthCredentials("root", "password"),
		}, nil},
		{"All Ops Together", []adapterOption{
			OpEndpoints("http://localhost:8529"),
			OpFieldMapping("p", "sub", "obj", "act"),
			OpDatabaseName("casbin"),
			OpCollectionName("casbin_rules_tests"),
			OpAutocreate(true),
			OpFieldMapping("p", "sub", "obj", "act"),
			OpBasicAuthCredentials("root", "password"),
		}, nil},
	}

	for _, tt := range operatorstests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewAdapter(tt.in...)
			if tt.expectedErr == nil && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if tt.expectedErr != nil && !tt.expectedErr(err) {
				t.Fatalf("Error other than expected: %v", err)
			}
		})
	}
}

func TestArangodbLoad(t *testing.T) {
	Convey("Given arangodb adapter", t, func() {
		ad, err := NewAdapter(
			OpFieldMapping("Type", "Arg0", "Arg1", "Arg2"),
			OpCollectionName("casbin_TestArangodbLoad"),
		)
		So(err, ShouldBeNil)

		Convey("And casbin enforcer using that adapter", func() {
			enforcer, err := newEnforcer()
			So(err, ShouldBeNil)
			enforcer.SetAdapter(ad)

			Reset(func() {
				err = truncateCollection(ad)
				So(err, ShouldBeNil)
			})

			Convey("When database is initialized with fixtures", func() {

				err = loadFixtures(ad, []string{
					"p,ADMIN,update,crazyBook",
					"p,ADMIN,truncate,crazyBook",
					"p,USER,insert,crazyBook",
				})
				So(err, ShouldBeNil)

				err = enforcer.LoadPolicy()
				So(err, ShouldBeNil)

				Convey("Enforcer should sucessfully enforce against all policies", func() {
					result, err := enforcer.Enforce("ADMIN", "update", "crazyBook")
					So(err, ShouldBeNil)
					So(result, ShouldBeTrue)

					result, err = enforcer.Enforce("ADMIN", "truncate", "crazyBook")
					So(err, ShouldBeNil)
					So(result, ShouldBeTrue)

					result, err = enforcer.Enforce("USER", "insert", "crazyBook")
					So(err, ShouldBeNil)
					So(result, ShouldBeTrue)
				})
			})
		})
	})
}

func TestArangodbSave(t *testing.T) {
	Convey("Given arangodb adapter", t, func() {
		ad, err := NewAdapter(
			OpFieldMapping("Type", "Arg0", "Arg1", "Arg2"),
			OpCollectionName("casbin_TestArangodbSave"),
		)
		So(err, ShouldBeNil)
		Convey("And casbin enforcer using that adapter", func() {
			enforcer, err := newEnforcer()
			So(err, ShouldBeNil)
			enforcer.SetAdapter(ad)

			Reset(func() {
				err = truncateCollection(ad)
				So(err, ShouldBeNil)
			})

			Convey("When policies are added and saved", func() {
				_, err = enforcer.AddPolicy("ADMIN", "write", "book")
				So(err, ShouldBeNil)
				_, err = enforcer.AddPolicy("USER", "read", "book")
				So(err, ShouldBeNil)
				_, err = enforcer.AddGroupingPolicy("adam", "ADMIN")
				So(err, ShouldBeNil)
				_, err = enforcer.AddGroupingPolicy("beata", "USER")
				So(err, ShouldBeNil)

				err := enforcer.SavePolicy()
				So(err, ShouldBeNil)

				Convey("Database should have policies saved", func() {
					content, err := getAllDbContent(ad)
					So(err, ShouldBeNil)
					So(content, ShouldResemble, map[string]bool{
						"p,ADMIN,write,book": true,
						"p,USER,read,book":   true,
						"g,adam,ADMIN":       true,
						"g,beata,USER":       true,
					})
				})
			})

			Convey("When policies are added and removed", func() {
				_, err = enforcer.AddPolicy("USER", "read", "emptyBook")
				So(err, ShouldBeNil)
				_, err = enforcer.AddPolicy("ADMIN", "write", "emptyBook")
				So(err, ShouldBeNil)

				err = enforcer.SavePolicy()
				So(err, ShouldBeNil)

				_, err = enforcer.RemovePolicy("ADMIN", "write", "emptyBook")
				So(err, ShouldBeNil)

				err = enforcer.SavePolicy()
				So(err, ShouldBeNil)

				Convey("Database should have policies saved", func() {
					content, err := getAllDbContent(ad)
					So(err, ShouldBeNil)
					So(content, ShouldResemble, map[string]bool{
						"p,USER,read,emptyBook": true,
					})
				})
			})
		})

		Convey("And casbin enforcer (using that adapter) with AUTOSAVE enabled", func() {
			enforcer, err := newEnforcer()
			So(err, ShouldBeNil)
			enforcer.SetAdapter(ad)
			enforcer.EnableAutoSave(true)

			Reset(func() {
				err = truncateCollection(ad)
				So(err, ShouldBeNil)
			})

			Convey("When policies are added", func() {
				_, err = enforcer.AddPolicy("ADMIN", "write", "emptyBook")
				So(err, ShouldBeNil)
				_, err = enforcer.AddPolicy("USER", "read", "emptyBook")
				So(err, ShouldBeNil)
				_, err = enforcer.AddGroupingPolicy("cezary", "USER")
				So(err, ShouldBeNil)
				_, err = enforcer.AddGroupingPolicy("diana", "ADMIN")
				So(err, ShouldBeNil)

				Convey("Database should have policies saved", func() {
					content, err := getAllDbContent(ad)
					So(err, ShouldBeNil)
					So(content, ShouldResemble, map[string]bool{
						"p,ADMIN,write,emptyBook": true,
						"p,USER,read,emptyBook":   true,
						"g,diana,ADMIN":           true,
						"g,cezary,USER":           true,
					})
				})
			})

			Convey("When policies are added then some removed", func() {
				_, err = enforcer.AddPolicy("USER", "read", "emptyBook")
				So(err, ShouldBeNil)
				_, err = enforcer.AddPolicy("ADMIN", "write", "emptyBook")
				So(err, ShouldBeNil)
				_, err = enforcer.RemovePolicy("ADMIN", "write", "emptyBook")
				So(err, ShouldBeNil)

				Convey("Database should have policies saved", func() {
					content, err := getAllDbContent(ad)
					So(err, ShouldBeNil)
					So(content, ShouldResemble, map[string]bool{
						"p,USER,read,emptyBook": true,
					})
				})
			})

			Convey("When policies are added then some removed with RemoveFilteredPolicy", func() {
				_, err = enforcer.AddPolicy("USER", "read", "emptyBook")
				So(err, ShouldBeNil)
				_, err = enforcer.AddPolicy("ADMIN", "write", "emptyBook")
				So(err, ShouldBeNil)
				_, err = enforcer.AddPolicy("ADMIN", "write", "plainBook")
				So(err, ShouldBeNil)
				_, err = enforcer.RemoveFilteredPolicy(2, "emptyBook")
				So(err, ShouldBeNil)

				Convey("Database should have policies saved", func() {
					content, err := getAllDbContent(ad)
					So(err, ShouldBeNil)
					So(content, ShouldResemble, map[string]bool{
						"p,ADMIN,write,plainBook": true,
					})
				})
			})

		})

	})
}

// ====== end of test cases ======

var rbacModel = `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
`

func newEnforcer() (*casbin.Enforcer, error) {
	m, err := model.NewModelFromString(rbacModel)
	if err != nil {
		return nil, err
	}
	e, err := casbin.NewEnforcer(m)
	if err != nil {
		return nil, err
	}
	return e, nil
}

type testPolicy struct {
	Type string
	Arg0 string
	Arg1 string
	Arg2 string
}

func newFromString(s string) testPolicy {
	ss := strings.Split(s, ",")
	tp := testPolicy{}
	if len(ss) >= 1 {
		tp.Type = ss[0]
	}
	if len(ss) >= 2 {
		tp.Arg0 = ss[1]
	}
	if len(ss) >= 3 {
		tp.Arg1 = ss[2]
	}
	if len(ss) >= 4 {
		tp.Arg2 = ss[3]
	}
	return tp
}

func (p *testPolicy) String() string {
	result := p.Type
	if p.Arg0 != "" {
		result = result + "," + p.Arg0
	}
	if p.Arg1 != "" {
		result = result + "," + p.Arg1
	}
	if p.Arg2 != "" {
		result = result + "," + p.Arg2
	}
	return result
}

func getAllDbContent(ad persist.Adapter) (map[string]bool, error) {
	a, ok := ad.(*adapter)
	if !ok {
		return nil, errors.New("Adapter is not arangodb.adapter type as expected")
	}

	query := fmt.Sprintf("FOR d IN %s LIMIT 100 RETURN d", a.collectionName)
	cursor, err := a.database.Query(context.Background(), query, nil)
	if err != nil {
		return nil, err
	}
	defer cursor.Close()
	result := make(map[string]bool)
	for {
		tp := testPolicy{}
		_, err := cursor.ReadDocument(context.Background(), &tp)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return nil, err
		}

		result[tp.String()] = true
	}
	return result, nil
}

func truncateCollection(ad persist.Adapter) error {
	a, ok := ad.(*adapter)
	if !ok {
		return errors.New("Adapter is not arangodb.adapter type as expected")
	}
	err := a.collection.Truncate(context.Background())
	return err
}

func loadFixtures(ad persist.Adapter, fixtures []string) error {
	a, ok := ad.(*adapter)
	if !ok {
		return errors.New("Adapter is not arangodb.adapter type as expected")
	}
	for _, line := range fixtures {
		testPolicy := newFromString(line)
		_, err := a.collection.CreateDocument(context.Background(), &testPolicy)
		if err != nil {
			return err
		}
	}
	return nil
}
