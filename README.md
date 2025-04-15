# Usage

Expects a Chrome browser to be started with remote debugging enabled.
e.g.  on MacOS

```bash
open -a "Google Chrome" --args --remote-debugging-port=9222
```


```bash
> samlsignature -h
Usage of samlsignature:
  -cert string
        [required] Path to certificate file
  -date string
        Date to use for validation (YYYY-MM-DD) (default current date)
  -postURL string
        IdP's Post URL (default "https://accounts.sap.com/saml2/idp/sso")
  -wayflessURL string
        Wayfless URL (default "https://dl.acm.org/action/ssostart?idp=https://accounts.sap.com")
```

## Sample usage

```bash
> samlsignature -cert acm.cert -wayflessURL "https://dl.acm.org/action/ssostart?idp=https://accounts.sap.com"  -postURL https://accounts.sap.com/saml2/idp/sso 
```
