# Butler Git Hooks

## Usage

1. Create a folder `git_hooks` in the root directory of your template.
2. Create git hook specific template
3. Run Butler and run `Install Git Hooks`

**pre-commit**
```sh
#!/bin/sh

echo "hook executed!"
```

_Git Hooks are installed automatically when a new project template is created._