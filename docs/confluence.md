# Butler confluence
Before you can use confluence commands you have to setup authentication & authorization
## Configure authentication

```
BUTLER_CONFLUENCE_URL=https://confluence.company.de     The base url of your confluence server (string, required)
BUTLER_CONFLUENCE_AUTH_METHOD=basic                     The authentication method (string, required)
BUTLER_CONFLUENCE_BASIC_AUTH=username,password          The basic authentication credentials comma seperated (string, required)
```

## Configure confluence permission
You have to assign each dev the global permission "Create Space(s)".
When a space is created, the creator automatically has the 'Admin' permission for that space and can perform space-wide administrative functions.