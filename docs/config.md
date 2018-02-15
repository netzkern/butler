# Butler config

In the config you can define which project templates are accessible for the user.

## The butler.yml file

```yml
templates:
  - name:   The template name (string, required)
    url:    The git url to the template (string, required)

variables:
  test:     The value for custom variable
```

### Custom variables

You can define custom variables to use them inside project templates. Custom
template variables have priority over local variables.

### Distribute config

You can set the environment variable `BUTLER_CONFIG_URL` to any url to load
your config from an external storage. This make it easy to distribute template
updates across a company. You local configuration is merged.
