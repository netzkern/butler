# Butler Template Surveys

Your are able to create an interactive survey before your template is proceed. The results can be used to build templates for user-specific requirements.

## How to create a survey?

1.  Create a config file `butler-survey.yml` in the root directory of your template repository.
2.  Create questions based on the [format](#configuration) below.
3.  Build your template with the [easy to use](/docs/templateSyntax.md#get-survey-results) template syntax.
4.  Add new template entry in your `butler.yml` file.
5.  Run butler and create a new project.

## The butler-survey.yml file

```
deprecated:     Whether or not this template is deprecated (optional, boolean)
butlerVersion:  The required butler version (optional, semver range string e.g "1.0.x" or ">1.0.0 <2.0.0 || >=3.0.0")

questions:
  - type:     The question type ([input, select, multiselect, password, confirm], required)
    name:     The indentifier of your question to access it in your template (string, required)
    message:  The question (string, required)
    options:  The available options only required for question types of select and multiselect ([]string, optional)
    default:  The default value ([]string for select otherwise string, optional)
    help:     The help message (string, optional)
    required: Whether or not this question is required (boolean, optional)

afterHooks:
  - name:     The command name (string, required)
    cmd:      The command to execute (string, required)
    args:     The arguments for the cmd ([]string, optional)
    verbose:  The command output is printend in the terminal (boolean, optional)
    enabled:  The template expression which has to be evaluated to `true` when `false` the command is skipped (string, optional)
    required: The command is required and will abort the hooks pipeline when it couldn't be executed successfully (boolean, optional)

variables:
  test:       The value for custom variable
```

## After hooks

Hooks are executed after the project is created. The hook pipeline is aborted when a command return an error which was marked as `required:true`.
The hook process will inherit all environment variables from the parent process.

You have access to the survey results inside your hooks. The results are exposed with environment variables.

```sh
BUTLER_<NAME>=a # for single values
BUTLER_<NAME>=a,b # for multiple values like "multiselect" question
```

## Custom variables

You can define custom variables. In case of a conflict the template variables have priority over local variables.
