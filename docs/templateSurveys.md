# Butler Template Surveys

Your are able to create an interactive survey before your template is proceed. The results can be used as template variables.

## How to create a survey?

1. Create a config file `butler-survey.yml` in the root directory of the template repository.
2. Create options based on the format below.
3. [Use](/docs/templateSyntax.md#get-survey-results) the answers in your template.
4. Run butler and select your template.

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
    name: languages
    message: "Choose your favorite programming languages:"
    options: ["c#", "go", "javascript"]
    default: ["c#"]
```

**Configuration**
- type: input, select, multiselect `string`
- name: the indentifier of your question to access it in your template `string`
- message: the question or statement `string`
- options: the choices if you use select or mulitselect questions `[]string`
- default: depends on the type `string` for select and `[]string` for multiselect questions
- required: `boolean`
- help: `string`
