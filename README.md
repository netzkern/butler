<p align="center">
<img src="https://raw.githubusercontent.com/netzkern/butler/master/logo.png" alt="butler" style="max-width:100%;">
</p>

Welcome to Butler, your personal assistent to scaffolding your projects.
Shipped with binaries for Mac, Win and Linux (64-bit).

# Usage

[Download here](https://github.com/netzkern/butler/releases)

```sh
$ butler
```

# Commands

## Project Templates

This command will create a new project template and replaces all variables in files and filenames. Project templates are managed in the `butler.yml` file. Butler is shipped with a default config. If you want to create a project template look [here](https://golang.org/pkg/text/template/) for the template features. We use a unique delimiter to avoid collsion with existing template engines.

### Delimiter

```
butler{ .Project.Name }
```

### Available variables:

- `Project.Name`: Project name
- `Project.Description`: Project description
- `Date`: Current Date (RFC3339)
- `Year`: Current year

You can specify custom variables in the `butler.yml` file. They can be accessed e.g `butler{ .Vars.company }`.

### Demo Template

[example-project-template](https://github.com/netzkern/example-project-template)

### Credits

<div>Icons made by <a href="http://www.freepik.com" title="Freepik">Freepik</a> from <a href="https://www.flaticon.com/" title="Flaticon">www.flaticon.com</a> is licensed by <a href="http://creativecommons.org/licenses/by/3.0/" title="Creative Commons BY 3.0" target="_blank">CC 3.0 BY</a></div>