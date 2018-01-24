# Butler Template Syntax

We use the template engine from Go. [Here](https://golang.org/pkg/text/template/) you can find an overview about all template features. We use a unique delimiter to avoid collsion with existing template engines.

## Delimiter

Inside files you have to use:
```
butler{<expr>} 
```
For directory or file names you have to use a different delimiter to save character because windows has a path [limit](https://msdn.microsoft.com/en-us/library/windows/desktop/aa365247(v=vs.85).aspx). We recommend to use short question names in the `survey-butler.yml`.
```
{<expr>} 
```


## Where can I use templates?
- Filenames
- Directories
- Text files (.html, .md, .txt, .cshtml, .cs, .js ...)

# Built in

## Project details
- `butler{ .Project.Name }` Return the project name
- `butler{ .Project.Description }` Return the project description
- `butler{ .Date }` Return the date (RFC3339)
- `butler{ .Year }` Return the year (4-digits)

## Define custom variables
You can specify custom variables in the `butler.yml` file.

```yaml
variables:
  company: netzkern
  email: info@netzkern.de
```
Usage:
```
butler{ .Vars.company }
```

## Helper functions
- `butler{ toCamelCase .Project.Name }` Return the project name in camel-case.
- `butler{ toPascalCase "foo-bar" }` Return the project name in pascal-case.
- `butler{ toSnakeCase "foo-bar" }` Return "foo-bar" in snake-case.
- `butler{ toPascalCase "foo-bar" }` Return "foo-bar" in pascal-case.
- `butler{ print uuid }` Return a random UUID Version 4.

## Define variables in templates
```
butler{$id := uuid} // generate id
butler{$id} // print id
```

## Get survey results
We generate getter functions to provide an easier access to survey results. If you set the name e.g to `name=db` to a question you can access the value with:

```
butler{getDb}
```
## Conditions in templates
```
butler{if eq getDb "mongodb"}
// your template
butler{end}
```

## Conditional directories and files
Based on the survey you can decide which directories or files should be included or removed. The following example will include the folder when the question about the `database` will be answered with `mongodb`.
```
Folder: {if eq getDb `mongodb` }mongodb{end}
```
Build the filename based on an answer:
```
Filename: {print getColor `.md`}
```
