package main

import (
	"log"
	"github.com/hashicorp/terraform/helper/schema"
	"crypto/rand"
	"fmt"
	"path/filepath"
	"os"
	"io/ioutil"
	"github.com/pkg/errors"
)

func resourceShellScript() *schema.Resource {
	return &schema.Resource{
		Create: resourceShellScriptCreate,
		Delete: resourceShellScriptDelete,
		Read:   resourceShellScriptRead,
		Update: resourceShellScriptUpdate,
		Schema: map[string]*schema.Schema{
			"command_directory" : {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					basePath := val.(string)
					createScript := filepath.Join(basePath, "create.sh")
					if _, err := os.Stat(createScript); os.IsNotExist(err) {
						errs = append(errs, fmt.Errorf("can't find required scripts in command_directory: %q", createScript))
					}
					deleteScript := filepath.Join(basePath, "delete.sh")
					if _, err := os.Stat(deleteScript); os.IsNotExist(err) {
						errs = append(errs, fmt.Errorf("can't find required scripts in command_directory: %q", deleteScript))
					}
					return
				},
			},
			"command_create": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},
			"command_read": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
			},
			"command_update": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
			},
			"command_delete": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},
			"idempotent": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default: false,
			},
			"environment": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     schema.TypeString,
			},
			"working_directory": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  ".",
			},
			"recreate_triggers": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
			},
			"update_trigger": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "created",
			},
			"output": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     schema.TypeString,
			},
		},
	}
}

func getScriptForAction(d *schema.ResourceData, action string) (string, error){
	resultScript := ""
	explicitCreateCommand := d.Get("command_create").(string)
	commandDirectory := d.Get("command_directory").(string)

	if explicitCreateCommand == "" && commandDirectory == "" {
		//error is invalid definition
		return "", errors.New("No explicit or inline commands are defined, invalid definition!")

	} else if commandDirectory != "" {
		var scriptErr error = nil
		//check for default, might fail depending on action
		_, defaultErr := getDefaultScriptForAction(action)
		//find actual script, might be missing then can use default
		resultScript, scriptErr = getDirectoryScriptForAction(d, action)
		//if we don't have a script defined, and it doesn't have a default then fail, otherwise set default
		if scriptErr != nil && defaultErr != nil {
			return "", scriptErr
		}

	} else {
		resultScript = d.Get("command_" + action).(string)
	}

	//if no script then use defaults
	if resultScript == "" {
		var defaultErr error = nil
		resultScript, defaultErr = getDefaultScriptForAction(action)
		if defaultErr != nil {
			return "", errors.New("No command defined or found or could be defaulted for action: " + action)
		}
	}
	return resultScript, nil
}

func getDirectoryScriptForAction(d *schema.ResourceData, action string) (string, error){
	commandDirectory := d.Get("command_directory").(string)
	scriptPath := filepath.Join(commandDirectory, action + ".sh")
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return "", fmt.Errorf("Can't find command from directory %q with name %q", commandDirectory, action + ".sh")
	}

	resultBytes, err := ioutil.ReadFile(scriptPath)
	if err != nil {
		//error is invalid definition
		return "", fmt.Errorf("Error reading command from directory %q with name %q", commandDirectory, action + ".sh")
	}
	return string(resultBytes[:]), nil
}

func getDefaultScriptForAction(action string) (string, error) {
	if action == "read" {
		return "IN=$(cat)\nprintf $IN >&3", nil
	} else if action == "read" {
		return "", nil
	} else {
		return "", errors.New("No default script for action")
	}
}

func resourceShellScriptCreate(d *schema.ResourceData, meta interface{}) error {
	command, commandErr := getScriptForAction(d, "create")
	if commandErr != nil {
		return commandErr
	}

	vars := d.Get("environment").(map[string]interface{})
	environment := readEnvironmentVariables(vars)
	workingDirectory := d.Get("working_directory").(string)

	//obtain exclusive lock
	shellMutexKV.Lock(shellScriptMutexKey)
	
	extraout, _, _, err := runCommand(command, "", environment, workingDirectory)
	if err != nil {
		return err
	}

	//try and parse extraout as JSON to expose, could just be a string
	output, err := parseJSON(extraout)
	if err != nil {
		log.Printf("[DEBUG] error parsing extraout into json: %v", err)
		output = make(map[string]string)
		d.Set("output", output)
	} else {
		d.Set("output", output)
	}

	//create random uuid for the id, changes to inputs will prompt update or recreate so it doesn't need to change
	idBytes := make([]byte, 16)
    _, randErr := rand.Read(idBytes)
    if randErr != nil {
        return randErr
    }
    d.SetId(hash(string(idBytes[0:])))

	//once creation has finished setup update_trigger so update works
	if len(extraout) > 0 {
		//if we have some content on extraout (used for diff) use that
		d.Set("update_trigger", extraout)
		shellMutexKV.Unlock(shellScriptMutexKey)
		return nil

	} else {
		shellMutexKV.Unlock(shellScriptMutexKey)
		//otherwise set update_trigger fix the read()
		return resourceShellScriptRead(d, meta)
	}
	
}

func resourceShellScriptRead(d *schema.ResourceData, meta interface{}) error {
	command, commandErr := getScriptForAction(d, "read")
	if commandErr != nil {
		return commandErr
	}

	vars := d.Get("environment").(map[string]interface{})
	environment := readEnvironmentVariables(vars)
	workingDirectory := d.Get("working_directory").(string)
	input := d.Get("update_trigger").(string)

	//obtain exclusive lock
	shellMutexKV.Lock(shellScriptMutexKey)
	defer shellMutexKV.Unlock(shellScriptMutexKey)

	extraout, _, _, err := runCommand(command, input, environment, workingDirectory)
	if err != nil {
		return err
	}
	output, err := parseJSON(extraout)
	if err != nil {
		log.Printf("[DEBUG] error parsing extraout into json: %v", err)
		var output map[string]string
		d.Set("output", output)
	} else {
		d.Set("output", output)
	}

	if len(extraout) == 0 {
		d.SetId("")
	} else {
		log.Printf("[DEBUG] Have update_trigger: " + input)
		log.Printf("[DEBUG] Setting as update_trigger: " + extraout)
		d.Set("update_trigger", extraout)
	}
	return nil
}

func resourceShellScriptUpdate(d *schema.ResourceData, meta interface{}) error {
	action := "update"
	//if command is idempotent then use the create command
	if d.Get("idempotent").(bool) == true {
		action = "create"
	}
	command, commandErr := getScriptForAction(d, action)
	if commandErr != nil {
		return commandErr
	}

	vars := d.Get("environment").(map[string]interface{})
	environment := readEnvironmentVariables(vars)
	workingDirectory := d.Get("working_directory").(string)
	input := d.Get("update_trigger").(string)

	_, _, _, err := runCommand(command, input, environment, workingDirectory)
	if err != nil {
		return err
	}

    return resourceShellScriptRead(d, meta)
}

func resourceShellScriptDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Deleting shell script resource")
	command, commandErr := getScriptForAction(d, "delete")
	if commandErr != nil {
		return commandErr
	}

	vars := d.Get("environment").(map[string]interface{})
	environment := readEnvironmentVariables(vars)
	workingDirectory := d.Get("working_directory").(string)
	input := d.Get("update_trigger").(string)
	//obtain exclusive lock
	shellMutexKV.Lock(shellScriptMutexKey)
	defer shellMutexKV.Unlock(shellScriptMutexKey)

	_, _, _, err := runCommand(command, input, environment, workingDirectory)
	if err != nil {
		return err
	}

	d.SetId("")
	return nil
}