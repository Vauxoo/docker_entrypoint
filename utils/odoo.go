package utils

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/ini.v1"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// GetOdooUser is for future use, so far we do not plan to use other user that odoo to execute the instance
func GetOdooUser() string {
	user := os.Getenv("ODOO_USER")
	if user == "" {
		user = "odoo"
	}
	return user
}

// GetConfigFile will return the odoo config file path that will be used by default
func GetConfigFile() string {
	fileEnv := os.Getenv("ODOO_CONFIG_FILE")
	if fileEnv != "" {
		return fileEnv
	}
	return "/home/odoo/.openerp_serverrc"
}

// GetInstanceType will use by default INSTANCE_TYPE env var because is the one we have been using in DeployV
// for over 5 years, but as Odoo added a similar one we use it too. It is important to notice that the values must match
// for example "updates" and "staging" matchm because are the same stage, but different name
func GetInstanceType() (string, error) {
	it := os.Getenv("INSTANCE_TYPE")
	ost := os.Getenv("ODOO_STAGE")
	switch {
	case ost == "" && it == "":
		return "", fmt.Errorf("cannot determine the instance type, env vars INSTANCE_TYPE and/or ODOO_STAGE 'must' be defined and match")
	case ost == "":
		return it, nil
	case it == "production" && ost == "production":
		return "production", nil
	case it == "updates" && ost == "staging":
		return "updates", nil
	case it == "develop" && ost == "dev":
		return "develop", nil
	case it == "test" && ost == "staging":
		return "test", nil
	}
	return "", fmt.Errorf("cannot determine the instance type, env vars INSTANCE_TYPE and ODOO_STAGE 'must' match, got: 'INSTANCE_TYPE=%s' and 'ODOO_STAGE=%s'", it, ost)

}

// FilterStrings receive a slice of strings and filter them by the 'odoorc_' prefix because these are the variables that
// will be replaced in the configuration file, returns them as a map of strings where the key is the key of the
// configuration
func FilterStrings(list []string) map[string]string {
	res := make(map[string]string)
	for _, v := range list {
		parts := strings.SplitN(v, "=", 2)
		if len(parts) < 2 {
			continue
		}
		if strings.HasPrefix(strings.ToLower(parts[0]), "odoorc_") {
			res[parts[0]] = parts[1]
		}
	}
	return res
}

// GetOdooVars receives a slice of strings, filter them by the prefix and return them ready to be used in the
// configuration file
func GetOdooVars(vars []string) map[string]string {
	res := make(map[string]string)
	list := FilterStrings(vars)
	for k, v := range list {
		key := strings.TrimPrefix(strings.ToLower(k), "odoorc_")
		res[key] = v
	}
	return res
}

// UpdateOdooConfig saves the ini object
func UpdateOdooConfig(config *ini.File) error {
	cfgFile := GetConfigFile()
	if err := config.SaveTo(cfgFile); err != nil {
		return err
	}
	return nil
}

// UpdateSentry check if sentry is enabled in such case adds/updates the values in the ini condiguration file
// setting the environment and the odoo instance path
func UpdateSentry(config *ini.File, instanceType string) {
	if !config.Section("options").HasKey("sentry_enabled") {
		return
	}
	sentryStr := config.Section("options").Key("sentry_enabled").Value()
	isEnabled, err := strconv.ParseBool(sentryStr)
	if err != nil {
		return
	}
	if isEnabled {
		config.Section("options").Key("sentry_odoo_dir").SetValue("/home/odoo/instance/odoo")
		config.Section("options").Key("sentry_environment").SetValue(instanceType)
	}
}

// SetupWorker will update the configuration to match the desired type of container, for example:
// if you wish to run a cron only container set the containerType parameter to cron and this func will disable the
// longpolling and the xmlrpc service
func SetupWorker(config *ini.File, containerType string) {
	switch strings.ToLower(containerType) {
	case "worker":
		config.Section("options").Key("odoorc_http_enable").SetValue("True")
		config.Section("options").Key("max_cron_threads").SetValue("0")
		config.Section("options").Key("workers").SetValue("0")
		config.Section("options").Key("xmlrpcs").SetValue("False")

	case "cron":
		config.Section("options").Key("odoorc_http_enable").SetValue("False")
		config.Section("options").Key("max_cron_threads").SetValue("1")
		config.Section("options").Key("workers").SetValue("0")
		config.Section("options").Key("xmlrpcs").SetValue("False")
		config.Section("options").Key("xmlrpc").SetValue("False")

	case "longpoll":
		config.Section("options").Key("odoorc_http_enable").SetValue("False")
		config.Section("options").Key("max_cron_threads").SetValue("0")
		config.Section("options").Key("workers").SetValue("2")
		config.Section("options").Key("xmlrpcs").SetValue("False")
	}
}

// UpdateFromVars will updae the odoo configuration from env vars wich should start with ODOORC_ prefix, if the exists
// the value  will be updated else the parameter will be added to the 'options' section which is the default for Odoo.
// If you wish to add it to another section add the desired section to '/external_files/openerp_serverrc' or add
// the file with only that section to '/external_files/odoocfg'
func UpdateFromVars(config *ini.File, odooVars map[string]string) {
	sections := config.Sections()
	for k, v := range odooVars {
		updated := false
		for _, section := range sections {
			if section.HasKey(k) {
				section.Key(k).SetValue(v)
				updated = true
				break
			}
		}
		// The key does not exist so we add it into the options section
		if !updated {
			config.Section("options").Key(k).SetValue(v)
		}
	}
}

// SetDefaults takes care of important defaults:
// - Won't allow admin as default super user password, a random string is generated
// - Won't allow to change the default ports because inside the container is not needed and will mess with the external
// - Disable logrotate since supervisor will handle that
func SetDefaults(config *ini.File) {
	config.Section("options").Key("xmlrpc_port").SetValue("8069")
	config.Section("options").Key("longpolling_port").SetValue("8072")
	config.Section("options").Key("logrorate").SetValue("False")
	if config.Section("options").Key("admin_passwd").Value() == "admin" ||
		config.Section("options").Key("admin_passwd").Value() == "" {
		config.Section("options").Key("admin_passwd").SetValue(RandStringRunes(64))
	}
}

// Odoo this func coordinates all the odoo configuration loading the config file, calling all the methods needed to
// update the configuration
func Odoo() error {
	log.Info("Preparing the configuration")

	if err := prepareFiles(); err != nil {
		return err
	}

	log.Info("Setting up the config file")
	odooCfg, err := ini.Load(GetConfigFile())
	if err != nil {
		log.Errorf("Error loading Odoo config: %s", err.Error())
		return err
	}

	odooVars := GetOdooVars(os.Environ())

	UpdateFromVars(odooCfg, odooVars)
	SetupWorker(odooCfg, os.Getenv("CONTAINER_TYPE"))
	instanceType, err := GetInstanceType()
	if err != nil {
		return err
	}
	log.Debugf("Instance type: %s", instanceType)
	UpdateSentry(odooCfg, instanceType)
	SetDefaults(odooCfg)
	autostart := true
	if os.Getenv("AUTOSTART") != "" {
		autostart, err = strconv.ParseBool(os.Getenv("AUTOSTART"))
		if err != nil {
			autostart = true
		}
		log.Debugf("Autostart: %v", autostart)
	}
	if err := UpdateAutostart(autostart, "/etc/supervisor/conf.d"); err != nil {
		return err
	}
	log.Info("Saving new Odoo configuration")
	if err := UpdateOdooConfig(odooCfg); err != nil {
		return err
	}

	return nil
}

func prepareFiles() error {

	if err := Copy("/external_files/openerp_serverrc", GetConfigFile()); err != nil {
		return err
	}

	if err := appendFiles(GetConfigFile(), "/external_files/odoocfg"); err != nil {
		return err
	}

	fsPath := os.Getenv("CONFIGFILE_PATH")
	if fsPath == "" {
		fsPath = "/home/odoo/.local/share/Odoo/filestore"
	}

	if _, err := os.Stat(fsPath); err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(fsPath, 0777)
			if err != nil {
				return err
			}
		}
	}

	cmds := []string{
		"chmod ugo+rwxt /tmp",
		"chmod ugo+rw /var/log/supervisor",
		fmt.Sprintf("chown odoo:odoo %s", filepath.Dir(fsPath)),
		fmt.Sprintf("chown odoo:odoo %s", fsPath),
		"chown -R odoo:odoo /home/odoo/.ssh",
	}

	for _, c := range cmds {
		log.Debugf("Running command: %s", c)
		if err := RunAndLogCmdAs(c, "", nil); err != nil {
			log.Errorf("Error running command: %s", err.Error())
			return err
		}
	}
	return nil
}
