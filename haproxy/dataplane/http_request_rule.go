package dataplane

import (
	"fmt"
	"net/http"

	"github.com/haproxytech/models/v2"
)

func (c *Dataplane) HTTPRequestRules(parentType, parentName string) ([]models.HTTPRequestRule, error) {
	type resT struct {
		Data []models.HTTPRequestRule `json:"data"`
	}

	var res resT

	err := c.makeReq(http.MethodGet, fmt.Sprintf("/v2/services/haproxy/configuration/http_request_rules?parent_type=%s&parent_name=%s", parentType, parentName), nil, &res)
	if err != nil {
		return nil, err
	}

	return res.Data, nil
}

func (t *tnx) CreateHTTPRequestRule(parentType, parentName string, rule models.HTTPRequestRule) error {
	if err := t.ensureTnx(); err != nil {
		return err
	}
	return t.client.makeReq(http.MethodPost, fmt.Sprintf("/v2/services/haproxy/configuration/http_request_rules?parent_type=%s&parent_name=%s&transaction_id=%s", parentType, parentName, t.txID), rule, nil)
}

func (c *Dataplane) HTTPResponseRules(parentType, parentName string) ([]models.HTTPResponseRule, error) {
	type resT struct {
		Data []models.HTTPResponseRule `json:"data"`
	}

	var res resT

	err := c.makeReq(http.MethodGet, fmt.Sprintf("/v2/services/haproxy/configuration/http_response_rules?parent_type=%s&parent_name=%s", parentType, parentName), nil, &res)
	if err != nil {
		return nil, err
	}

	return res.Data, nil
}

func (t *tnx) CreateHTTPResponseRule(parentType, parentName string, rule models.HTTPResponseRule) error {
	if err := t.ensureTnx(); err != nil {
		return err
	}
	return t.client.makeReq(http.MethodPost, fmt.Sprintf("/v2/services/haproxy/configuration/http_response_rules?parent_type=%s&parent_name=%s&transaction_id=%s", parentType, parentName, t.txID), rule, nil)
}
