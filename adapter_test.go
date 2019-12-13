// Copyright 2018 The casbin Authors. All Rights Reserved.
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
	"os"
	"reflect"
	"testing"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
)

func skipIfNonLocal(t *testing.T) {
	if os.Getenv("LOCAL") == "" {
		t.Skip("Tests skipped for non-local run")
	}
}

var operatorstests = []struct {
	name string
	in   []adapterOption
	out  error
}{
	{"Custom Endpoint", []adapterOption{OpEndpoints("http://localhost:8529")}, nil},
	{"Custom Database Name", []adapterOption{OpDatabaseName("casbin")}, nil},
	{"Custom Collection Name", []adapterOption{OpCollectionName("casbin_rules")}, nil},
	{"Custom Field Mapping", []adapterOption{OpFieldMapping("p", "sub", "obj", "act")}, nil},
	{"All Ops Together", []adapterOption{
		OpEndpoints("http://localhost:8529"),
		OpFieldMapping("p", "sub", "obj", "act"),
		OpDatabaseName("casbin"),
		OpCollectionName("casbin_rules_tests"),
		OpFieldMapping("p", "sub", "obj", "act")}, nil},
}

func TestArangodbNewAdapter(t *testing.T) {
	skipIfNonLocal(t)
	for _, tt := range operatorstests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewAdapter(tt.in...)
			if err != tt.out {
				t.Fatalf("Unexpected result: %v", err)
			}
		})
	}
}

func TestArangodbSaveAndLoadPolicies(t *testing.T) {
	skipIfNonLocal(t)
	e1 := prepareAndInitEnforcerUT(t, prepareAdapter(t))
	e2 := prepareEnforcerUT(t, prepareAdapter(t))

	err := e1.SavePolicy()
	if err != nil {
		t.Fatalf("Could not create adapter: %v", err)
	}

	e2.ClearPolicy()
	err = e2.LoadPolicy()
	if err != nil {
		t.Fatalf("Could not create adapter: %v", err)
	}

	p1 := e1.GetModel()["p"]["p"].Policy
	p2 := e2.GetModel()["p"]["p"].Policy
	g1 := e1.GetModel()["g"]["g"].Policy
	g2 := e2.GetModel()["g"]["g"].Policy

	if !reflect.DeepEqual(p1, p2) {
		t.Fatalf("Saved and loaded policies are not equal: %#v & %#v", p1, p2)
	}

	if !reflect.DeepEqual(g1, g2) {
		t.Fatalf("Saved and loaded group policies are not equal: %#v & %#v", g1, g2)
	}
}

func TestArangodbAutoAddAndRemovePolicies(t *testing.T) {
	skipIfNonLocal(t)
	e1 := prepareAndInitEnforcerUT(t, prepareAdapter(t))
	e1.EnableAutoSave(true)
	e1.ClearPolicy()
	r := addPolicy(t, e1, "xavier", "aaa", "write")
	if !r {
		t.Errorf("Could not auto add policy")
	}
	r = removePolicy(t, e1, "xavier", "aaa", "write")
	if !r {
		t.Errorf("Could not auto remove policy")
	}
}

func TestArangodbRemoveFilteredPolicies(t *testing.T) {
	skipIfNonLocal(t)
	e1 := prepareAndInitEnforcerUT(t, prepareAdapter(t))
	addPolicy(t, e1, "xavier", "aaa", "write")
	addPolicy(t, e1, "yvette", "aaa", "write")
	addPolicy(t, e1, "zuzanna", "aaa", "write")
	err := e1.SavePolicy()
	if err != nil {
		t.Fatalf("Could not add policy")
	}
	e1.EnableAutoSave(true)
	_, err = e1.RemoveFilteredPolicy(1, "aaa")
	if err != nil {
		t.Fatalf("Could not remove filtered policy")
	}
	e1.LoadPolicy()

	e2 := prepareAndInitEnforcerUT(t, nil)

	p1 := e1.GetModel()["p"]["p"].Policy
	p2 := e2.GetModel()["p"]["p"].Policy
	g1 := e1.GetModel()["g"]["g"].Policy
	g2 := e2.GetModel()["g"]["g"].Policy

	if !reflect.DeepEqual(p1, p2) {
		t.Fatalf("Saved and loaded policies are not equal: %#v & %#v", p1, p2)
	}

	if !reflect.DeepEqual(g1, g2) {
		t.Fatalf("Saved and loaded group policies are not equal: %#v & %#v", g1, g2)
	}
}

func addPolicy(t *testing.T, e *casbin.Enforcer, policy ...string) bool {
	r, err := e.AddPolicy(policy)
	if err != nil {
		t.Fatalf("Could not add policy: %v", err)
	}
	return r
}

func removePolicy(t *testing.T, e *casbin.Enforcer, policy ...string) bool {
	r, err := e.RemovePolicy(policy)
	if err != nil {
		t.Fatalf("Could not remove policy: %v", err)
	}
	return r
}

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

func prepareEnforcerUT(t *testing.T, a persist.Adapter) *casbin.Enforcer {
	m, err := model.NewModelFromString(rbacModel)
	if err != nil {
		t.Fatalf("Could not create casbin model: %v", err)
	}
	e, err := casbin.NewEnforcer(m)
	if err != nil {
		t.Fatalf("Could not create casbin enforcer: %v", err)
	}
	if a != nil {
		e.SetAdapter(a)
	}
	e.EnableAutoSave(false)

	return e
}

func prepareAndInitEnforcerUT(t *testing.T, a persist.Adapter) *casbin.Enforcer {
	e := prepareEnforcerUT(t, a)

	for _, v := range [][]interface{}{
		{"ADMIN", "read", "book1"},
		{"ADMIN", "write", "book1"},
		{"USER", "read", "book2"},
		{"GUEST", "read", "book3"},
	} {
		_, err := e.AddPolicy(v...)
		if err != nil {
			t.Fatal("Error adding policy")
		}
	}

	for _, v := range [][]interface{}{
		{"anastazja", "ADMIN"},
		{"urszula", "USER"},
		{"genowefa", "GUEST"},
	} {
		_, err := e.AddGroupingPolicy(v...)
		if err != nil {
			t.Fatal("Error adding grouping policy")
		}
	}

	return e
}

func prepareAdapter(t *testing.T) persist.Adapter {
	a, err := NewAdapter(OpCollectionName("casbin_tests"))
	if err != nil {
		t.Fatalf("Could not create adapter: %s", err.Error())
	}
	return a
}
