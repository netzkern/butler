# Butler Template Surveys

Your are able to create an interactive survey before your template is proceed. The results can be used to build templates for user-specific requirements.

## How to create a survey?

1. Create a config file `butler-survey.yml` in the root directory of your template repository.
2. Create questions based on the [format](#configuration) below.
3. Build your template with the [easy to use](/docs/templateSyntax.md#get-survey-results) template syntax.
4. Run butler and create your a new project.

**butler-survey.yml**
```yml
questions:
  - type: input
    name: drink
    message: "What is your favorite drink?"
    help: "Allowed character 0-9, A-Z, _-"
    required: true
  - type: select
    name: color
    message: "Choose a color:"
    options: ["red", "green", "yellow"]
  - type: multiselect
    name: lang
    message: "Choose your programming language:"
    options: ["c#", "go", "javascript"]
    default: ["c#"]
  - type: select
    name: db
    message: "Choose your database:"
    options: ["mongodb", "mssql", "redis"]
    required: true
  - type: password
    name: dbPassword
    message: "Please enter a db password"
    required: true
  - type: confirm
    name: printNode
    message: "Should we print the Node Version?"
    
afterHooks:
  - cmd: node
    args: ["-v"]
    enabled: getPrintNode
  - cmd: npm
    args: ["-v"]
    enabled: eq getDb "mongodb"
```

### Configuration

#### Questions
- type: input, select, multiselect, password, confirm `string`
- name: the indentifier of your question to access it in your template `string`
- message: the question or statement `string`
- options: the choices if you use select or mulitselect questions `[]string`
- default: depends on the type `string` for select and `[]string` for multiselect questions
- required: `boolean`
- help: `string`

#### After hooks
Hooks are executed after the project is created. The hook pipeline is aborted when any command return an error.

- cmd: command to execute `string`
- args: arguments `[]string`
- enabled: a template expression which has to be evaluated to `true` when `false` the command is skipped

## Access survey results in hook scripts

All results are available via environment variables:
```
BUTLER_DRINK=dwe
BUTLER_COLOR=red
BUTLER_LANG=c#,go,javascript
BUTLER_DB=mongodb
```
