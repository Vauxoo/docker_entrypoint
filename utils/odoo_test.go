package utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterStrings(t *testing.T) {
	values := []struct {
		input    []string
		expected map[string]string
	}{
		{
			[]string{"notfilter=asd", "odoorc_filtered=123", "CAPS_NOTFILTERED=wer", "ODOORC_INCAPS=caps"},
			map[string]string{"odoorc_filtered": "123", "ODOORC_INCAPS": "caps"},
		},
	}
	for _, v := range values {
		res := FilterStrings(v.input)
		if !assert.Equal(t, v.expected, res) {
			t.Errorf("Got: %+v, expected: %+v", res, v.expected)
		}
	}

}

func TestFilterGetOdooVars(t *testing.T) {
	values := []struct {
		input    []string
		expected map[string]string
	}{
		{
			[]string{"odoorc_filtered=123", "ODOORC_INCAPS=caps"},
			map[string]string{"filtered": "123", "incaps": "caps"},
		},
	}
	for _, v := range values {
		res := GetOdooVars(v.input)
		if !assert.Equal(t, v.expected, res) {
			t.Errorf("Got: %+v, expected: %+v", res, v.expected)
		}
	}
}

func TestGetOdooUser(t *testing.T) {
	res := GetOdooUser()
	assert.Equal(t, "odoo", res)
}

func TestGetConfigFile(t *testing.T) {
	res := GetConfigFile()
	assert.Equal(t, "/home/odoo/.openerp_serverrc", res)
	err := os.Setenv("ODOO_CONFIG_FILE", "/etc/odoo.conf")
	assert.NoError(t, err)
	res = GetConfigFile()
	assert.Equal(t, "/etc/odoo.conf", res)
	os.Unsetenv("ODOO_CONFIG_FILE")
}

func TestGetInstanceType(t *testing.T) {
	_, err := GetInstanceType()
	assert.Errorf(t, err, "cannot determine the instance type, env vars INSTANCE_TYPE and/or ODOO_STAGE 'must' be defined and match")

	err = os.Setenv("INSTANCE_TYPE", "test")
	assert.NoError(t, err)
	res, err := GetInstanceType()
	assert.NoError(t, err)
	assert.Equal(t, "test", res)

	err = os.Setenv("ODOO_STAGE", "dev")
	assert.NoError(t, err)
	_, err = GetInstanceType()
	assert.Errorf(t, err, "cannot determine the instance type, env vars INSTANCE_TYPE and ODOO_STAGE 'must' match")
}

//func TestUpdateSentry(t *testing.T) {
//	values := []struct{
//		input map[string]string
//		instanceType string
//		expected map[string]string
//	}{
//		{
//			map[string]string{"sentry_enabled": "true"},
//			"develop",
//			map[string]string{"sentry_enabled": "true", "sentry_odoo_dir": "/home/odoo/instance/odoo", "sentry_environment": "develop"},
//		},
//		{
//			map[string]string{"sentry_enabled": "false"},
//			"test",
//			map[string]string{"sentry_enabled": "false"},
//		},
//		{
//			map[string]string{"sentry_enabled": "True"},
//			"production",
//			map[string]string{"sentry_enabled": "True", "sentry_odoo_dir": "/home/odoo/instance/odoo", "sentry_environment": "production"},
//		},
//		{
//			map[string]string{"not_sentry": "True"},
//			"production",
//			map[string]string{"not_sentry": "True"},
//		},
//	}
//	for _, v := range values {
//		UpdateSentry(v.input, v.instanceType)
//		if !assert.Equal(t, v.expected, v.input) {
//			t.Errorf("Got: %+v, expected: %+v", v.input, v.expected)
//		}
//	}
//}
