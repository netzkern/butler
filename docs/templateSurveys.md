# Butler Template Surveys

Your are able to create an interactive survey before your template is proceed. The results can be used as template variables.

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
- name: the id of your question `string`
- message: the question `string`
- options: `[]string`
- default: depends on the type `string` or `[]string`
- required: `boolean`
- help: `string`
