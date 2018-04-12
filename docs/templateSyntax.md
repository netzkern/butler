# Butler Template Syntax

We use the template engine from Go. [Here](https://golang.org/pkg/text/template/) you can find an overview about all template features. We use a unique delimiter to avoid collsion with existing template engines.

## Delimiter

Inside files you have to use:

```
butler{<expr>}
```

For directory or file names you have to use a different delimiter to save character because windows has a path [limit](<https://msdn.microsoft.com/en-us/library/windows/desktop/aa365247(v=vs.85).aspx>). We recommend to use short question names in the `survey-butler.yml`.

```
{<expr>}
```

## Where can I use templates?

* Filenames
* Directories
* Template variables
* Text files (.html, .md, .txt, .cshtml, .cs, .js ...)

_Butler maintain a [list](https://github.com/netzkern/butler/blob/master/commands/template/binary_extensions.go) of extensions of binary files and disallow the parsing of these files_

# Built in

## Project details

* `butler{ .Project.Name }` Return the project name
* `butler{ .Project.Description }` Return the project description
* `butler{ .Date }` Return the date (RFC3339)
* `butler{ .Year }` Return the year (4-digits)

## Trim spaces around template actions

```go
{{23 -}} // remove trailing space
   <
{{- 45}} // remove leading space
```

formats as `23<45`

## Helper functions

### String

* `butler{ toCamelCase $string }` Convert argument to camelCase style string If argument is empty, return itself.
* `butler{ toPascalCase $string }` Convert argument to PascalCase style string If argument is empty, return itself.
* `butler{ toSnakeCase $string }` Convert argument to snake_case style string. If argument is empty, return itself.
* `butler{ toLowerCase $string }` Convert argument to lowercase style string If argument is empty, return itself.
* `butler{ toUpperCase $string }` Convert argument to UPPERCASE style string If argument is empty, return itself.
* `butler{ join $array $seperator }` Join concatenates the elements of a to create a single string.
* `butler{ replace $old $new $limit }` Replace returns a copy of the string s with the first n non-overlapping instances of old replaced by new.
* `butler{ contains $string $substring }` Contains reports whether substr is within s.
* `butler{ index $string $substring }` Contains reports whether substr is within s.
* `butler{ repeat $string $count }` Repeat returns a new string consisting of count copies of the string s.
* `butler{ split $string $sep }` Split slices s into all substrings separated by sep and returns a slice of the substrings between those separators.

### Path

* `butler{ joinPath $path, $path2 }` Joins any number of path elements into a single path, adding a Separator if necessary.
* `butler{ relPath $path }` Returns a relative path that is lexically equivalent to targpath when joined to basepath with an intervening separator.
* `butler{ absPath $path }` Returns an absolute representation of path.
* `butler{ basePath $path }` Returns the last element of path.
* `butler{ extPath $path }` Returns the file name extension used by path.

### Regex

* `butler{ (regex "[0-9]+").FindString "I'm 26 years old" }` Returns `26`
* For more methods look in the official [`Regex documentation`](https://golang.org/pkg/regexp/.)

### Generators

* `butler{ uuid }` Returns a random UUID Version 4 string.
* `butler{ randomInt $min $max }` Returns a random int between min and max

### Environment

* `butler{ cwd }` Returns the absolute path of the working directory.
* `butler{ env "name" }` Returns the value of the environment variable.

_All functions are written in camelCase_

## Define variables in templates

```
butler{$id := uuid} // generate id
butler{$id} // print id
```

## Access to custom variables

Custom variables can be defined in the local `butler.yml` file or in the template `butler-survey.yml` file.

```
butler{ .Vars.company }
```

## Interpolate custom variables

You have access to the survey result as well as template helper to define variables. This is useful if you want to work with short names in file or directory names.

```yml
variables:
  email: test@email.de
  projectName: "{ toUpperCase .Project.Name }"
  dbName: "{ toPascalCase getDb }"
  emailLowerCase: "{ toLowerCase .Vars.email }"
```

Inside your template
```
butler{ .Vars.projectName }
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
