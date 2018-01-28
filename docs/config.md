# Butler config

In the config you can define which project templates are accessible for the user.

**butler.yml**
```yml
templates:
  - name: Example 1
    url: https://github.com/netzkern/example-project-template.git
  - name: Example 2
    url: https://github.com/netzkern/example-project-template.git

variables:
  company: netzkern
  email: info@netzkern.de
```

### Configuration

#### Templates
- name: unique template name `string`
- url: git repository url `string`

#### Custom variables
You can also define custom variables which can be used inside project templates. Custom template variables have priority over local variables.

### Distribute your config
You can set the environment variable `BUTLER_CONFIG_URL` to url to load your config from an external storage. This make it easy to distribute template updates across a company.