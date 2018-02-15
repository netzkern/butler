# Butler Template Surveys

Your are able to create an interactive survey before your template is proceed.
The results can be used to build templates for user-specific requirements.

## How to create a survey?

1. Create a config file `butler-survey.yml` in the root directory of your
   template repository.
2. Create questions based on the [format](#configuration) below.
3. Build your template with the [easy to use](/docs/templateSyntax.md#get-survey-results)
   template syntax.
4. Run butler and create a new project.

## The butler-survey.yml file

```plain
deprecated:     Whether or not this template is deprecated (optional, boolean)
butlerVersion:  The required butler version (optional, semver range string e.g "1.0.x" or ">1.0.0 <2.0.0 || >=3.0.0")

questions:
  - type:     The question type ([input, select, multiselect, password, confirm], required)
    name:     The indentifier of your question to access it in your template (string, required)
    message:  The question (string, required)
    options:  The available options only required for question types of select and multiselect ([]string)
    default:  The default value ([]string for select otherwise string)
    help:     The help message (string, optional)
    required: Whether or not this question is required (boolean)

afterHooks:
  - cmd:      The command to execute (string, required)
    args:     The arguments for the cmd ([]string, optional)
    enabled:  The template expression which has to be evaluated to `true` when `false` the command is skipped (string, optional)

variables:
  test:       The value for custom variable
```

### After hooks

Hooks are executed after the project is created. The hook pipeline is aborted
when any command return an error.

### Custom variables

You can define custom variables. In case of a conflict the template variables
have priority over local variables.
