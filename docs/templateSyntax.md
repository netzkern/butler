# Butler Template Syntax

We use the template engine from Go. [Here](https://golang.org/pkg/text/template/) you can find an overview about all template features you can use in your templates. We use a unique delimiter to avoid collsion with existing template engines.

## Delimiter

```
butler{<expr>} 
```

## Where can I use templates?
- Filenames
- Directories
- Text files (.html, .md, .txt, .cshtml, .cs, .js ...)

## Custom variables
You can specify custom variables in the `butler.yml` file. They can be accessed by `butler{ .Vars.company }`.

```yaml
variables:
  company: netzkern
  email: info@netzkern.de
```

## Default variables
- Project name: butler{ .Project.Name }
- Project Description: butler{ .Project.Description }
- Current Date: butler{ .Date }
- Current Year: butler{ .Year }

## Custom variables
- Company: butler{ .Vars.company }
- Email: butler{ .Vars.email }
## Helper functions
- butler{ toCamelCase .Project.Name }
- butler{ toPascalCase "foo-bar" }
- butler{ toSnakeCase "foo-bar" }

## Get survey results
```
butler{getSurveyResult "database"}
```

## Conditions
```
butler{if eq (getSurveyResult "database") "mongodb"}

butler{end}
```
