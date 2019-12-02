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

var (
	ErrTooManyArguments      error = errors.New("policy has too many arguments")
	ErrInvalidPolicyDocument error = errors.New("db document does not match valid policy")
)

var defaultMapping []string = []string{"PType", "V0", "V1", "V2", "V3", "V4", "V5"}

type adapter struct {
	endpoints      []string
	mapping        []string
	collectionName string
	database       arango.Database
	query          string
	collection     arango.Collection
}

type adapterOption func(*adapter)

func OpEndpoints(endpoints ...string) func(*adapter) {
	return func(a *adapter) {
		a.endpoints = make([]string, 0, len(endpoints))
		for _, e := range endpoints {
			a.endpoints = append(a.endpoints, e)
		}
	}
}

func OpCollectionName(collectionName string) func(*adapter) {
	return func(a *adapter) {
		a.collectionName = collectionName
	}
}

func OpFieldMapping(mapping ...string) func(*adapter) {
	return func(a *adapter) {
		a.mapping = mapping
	}
}

func NewAdapter(options ...adapterOption) (persist.Adapter, error) {
	a := adapter{}
	a.collectionName = "casbin_rules"
	a.mapping = defaultMapping
	a.endpoints = []string{"http://127.0.0.1:8529"}

	for _, option := range options {
		option(&a)
	}

	conn, err := http.NewConnection(http.ConnectionConfig{
		Endpoints: a.endpoints,
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
	a.database = db

	var queryResult []string = make([]string, 0, len(a.mapping))
	for _, v := range a.mapping {
		queryResult = append(queryResult, `"`+v+`":d.`+v)
	}

	a.query = fmt.Sprintf("FOR d IN %s RETURN {%s}", a.collectionName, strings.Join(queryResult, ","))

	exists, err := db.CollectionExists(nil, a.collectionName)
	if err != nil {
		return nil, err
	}
	if !exists {
		_, err := db.CreateCollection(nil, a.collectionName, nil)
		// 1207 is ERROR_ARANGO_DUPLICATE_NAME - driver has no symbolic wrapper for it for now
		// ignores error that may happen if collection has been created in the meantime
		if err != nil && arango.IsArangoErrorWithErrorNum(err, 1207) {
			return nil, err
		}
	}

	col, err := db.Collection(nil, a.collectionName)
	if err != nil {
		return nil, err
	}
	a.collection = col
	return &a, nil
}

func (a *adapter) loadPolicyLine(line ArangoRule, model model.Model) error {
	key := line[a.mapping[0]]
	if key == "" {
		return ErrInvalidPolicyDocument
	}
	sec := key[:1]

	tokens := []string{}

	for _, name := range a.mapping[1:] {
		value, ok := line[name]
		if !ok || value == "" {
			break
		}
		tokens = append(tokens, value)
	}
	if len(tokens) == 0 {
		return ErrInvalidPolicyDocument
	}

	model[sec][key].Policy = append(model[sec][key].Policy, tokens)
	return nil
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
		err = a.loadPolicyLine(doc, model)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *adapter) savePolicyLine(ptype string, rule []string) (ArangoRule, error) {
	if 1+len(rule) > len(a.mapping) {
		return nil, ErrTooManyArguments
	}
	ruleList := make(ArangoRule, len(a.mapping))
	ruleList[a.mapping[0]] = ptype
	for i, v := range rule {
		ruleList[a.mapping[i+1]] = v
	}
	return ruleList, nil
}

// SavePolicy saves policy to database.
func (a *adapter) SavePolicy(model model.Model) error {
	var lines []interface{}

	for ptype, ast := range model["p"] {
		for _, rule := range ast.Policy {
			line, err := a.savePolicyLine(ptype, rule)
			if err != nil {
				return err
			}
			lines = append(lines, &line)
		}
	}

	for ptype, ast := range model["g"] {
		for _, rule := range ast.Policy {
			line, err := a.savePolicyLine(ptype, rule)
			if err != nil {
				return err
			}
			lines = append(lines, &line)
		}
	}
	err := a.collection.Truncate(nil)
	if err != nil {
		return err
	}
	_, _, err = a.collection.CreateDocuments(nil, lines)
	return err
}

// AddPolicy adds a policy rule to the storage.
func (a *adapter) AddPolicy(sec string, ptype string, rule []string) error {
	line, err := a.savePolicyLine(ptype, rule)
	if err != nil {
		return err
	}
	_, err = a.collection.CreateDocument(nil, line)
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
