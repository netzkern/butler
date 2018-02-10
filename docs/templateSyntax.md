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

_Butler maintain a [list](https://github.com/netzkern/butler/blob/master/commands/template/binary_extensions.go) of extensions from binary files and disallow the parsing of these files_

# Built in

## Project details
- `butler{ .Project.Name }` Return the project name
- `butler{ .Project.Description }` Return the project description
- `butler{ .Date }` Return the date (RFC3339)
- `butler{ .Year }` Return the year (4-digits)

## Access to custom variables
Custom variables can be defined in the local `butler.yml` file or in the template `butler-survey.yml` file.

```
butler{ .Vars.company }
```

## Helper functions
- `butler{ toCamelCase .Project.Name }` Transform a string to camel-case.
- `butler{ toPascalCase "foo-bar" }` Transform a string to pascal-case.
- `butler{ toSnakeCase "foo-bar" }` Transform a string to snake-case.
- `butler{ toPascalCase "foo-bar" }` Transform a string to pascal-case.
- `butler{ join $array "," }` Joins all elements of an array into a string and returns this string.
- `butler{ uuid }` Returns a random UUID Version 4 string.

_All functions are written in camelCase_

## Define variables in templates
```
butler{$id := uuid} // generate id
butler{$id} // print id
```

## Get survey results
We generate getter functions to provide an easier access to survey results. If you have a question with the name `color` the result is accessible by:

```
butler{ getColor }
```

## Get survey question
You can access the survey questions with the same approach
```
butler{ getColorQuestion }
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
