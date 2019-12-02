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
	// "errors"
	// "runtime"
	"errors"
	"fmt"
	"strings"

	arango "github.com/arangodb/go-driver"
	http "github.com/arangodb/go-driver/http"

	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
)

type ArangoRule map[string]string

var ErrTooManyArguments error = errors.New("policy has too many arguments")

var defaultMapping []string = []string{"PType", "V0", "V1", "V2", "V3", "V4", "V5"}

func newRule(policyType string, values ...string) (ArangoRule, error) {
	if 1+len(values) > len(defaultMapping) {
		return nil, ErrTooManyArguments
	}
	rule := make(ArangoRule, len(defaultMapping))
	rule[defaultMapping[0]] = policyType
	for i, v := range values {
		rule[defaultMapping[i+1]] = v
	}
	return rule, nil
}

type adapter struct {
	mapping    []string
	database   arango.Database
	query      string
	collection arango.Collection
}

func NewAdapter(urls []string) (persist.Adapter, error) {
	return NewAdapterWithMapping(urls, defaultMapping)
}

func NewAdapterWithMapping(urls []string, mapping []string) (persist.Adapter, error) {
	conn, err := http.NewConnection(http.ConnectionConfig{
		Endpoints: urls,
	})
	if err != nil {
		return nil, err
	}
	c, err := arango.NewClient(
		arango.ClientConfig{
			Connection: conn,
		},
	)
	if err != nil {
		return nil, err
	}
	db, err := c.Database(nil, "casbin")
	if err != nil {
		return nil, err
	}

	var queryResult []string = make([]string, 0, len(mapping))
	for _, v := range mapping {
		queryResult = append(queryResult, `"`+v+`":d.`+v)
	}

	col, err := db.Collection(nil, "casbin_rules")
	if err != nil {
		return nil, err
	}
	return &adapter{
		mapping:    mapping,
		database:   db,
		collection: col,
		query:      fmt.Sprintf("FOR d IN %s RETURN {%s}", "casbin_rules", strings.Join(queryResult, ",")),
	}, nil
}

func (a *adapter) loadPolicyLine(line ArangoRule, model model.Model) {
	key := line[a.mapping[0]]
	sec := key[:1]

	tokens := []string{}

	for _, name := range a.mapping[1:] {
		value, ok := line[name]
		if !ok || value == "" {
			break
		}
		tokens = append(tokens, value)
	}

	model[sec][key].Policy = append(model[sec][key].Policy, tokens)
}

// LoadPolicy loads policy from database.
func (a *adapter) LoadPolicy(model model.Model) error {
	cursor, err := a.database.Query(nil, a.query, nil)
	if err != nil {
		return err
	}
	defer cursor.Close()

	for {
		var doc map[string]string = make(map[string]string)
		_, err := cursor.ReadDocument(nil, &doc)
		if arango.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return err
		}
		a.loadPolicyLine(doc, model)
	}
	return nil
}

func savePolicyLine(ptype string, rule []string) ArangoRule {
	ruleList, _ := newRule(ptype, rule...)
	return ruleList
}

// SavePolicy saves policy to database.
func (a *adapter) SavePolicy(model model.Model) error {
	err := a.collection.Truncate(nil)
	if err != nil {
		return err
	}

	var lines []interface{}

	for ptype, ast := range model["p"] {
		for _, rule := range ast.Policy {
			line := savePolicyLine(ptype, rule)
			lines = append(lines, &line)
		}
	}

	for ptype, ast := range model["g"] {
		for _, rule := range ast.Policy {
			line := savePolicyLine(ptype, rule)
			lines = append(lines, &line)
		}
	}
	_, _, err = a.collection.CreateDocuments(nil, lines)
	return err
}

// AddPolicy adds a policy rule to the storage.
func (a *adapter) AddPolicy(sec string, ptype string, rule []string) error {
	line := savePolicyLine(ptype, rule)
	_, err := a.collection.CreateDocument(nil, line)
	return err
}

// RemovePolicy removes a policy rule from the storage.
func (a *adapter) RemovePolicy(sec string, ptype string, rule []string) error {
	return nil
}

// RemoveFilteredPolicy removes policy rules that match the filter from the storage.
func (a *adapter) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	return nil
}
