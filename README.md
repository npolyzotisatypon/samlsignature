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

POST request to: https://accounts.sap.com/saml2/idp/sso
SAMLRequest: <?xml version="1.0" encoding="UTF-8"?>
<urn:AuthnRequest AssertionConsumerServiceURL="https://dl.acm.org/action/saml2post" Destination="https://accounts.sap.com/saml2/idp/sso" ID="_-4275728805139421773" IssueInstant="2025-04-15T11:17:35.827Z" Version="2.0" xmlns:urn="urn:oasis:names:tc:SAML:2.0:protocol"><urn1:Issuer xmlns:urn1="urn:oasis:names:tc:SAML:2.0:assertion">https://dl.acm.org/shibboleth</urn1:Issuer><ds:Signature xmlns:ds="http://www.w3.org/2000/09/xmldsig#"><ds:SignedInfo><ds:CanonicalizationMethod Algorithm="http://www.w3.org/2001/10/xml-exc-c14n#"/><ds:SignatureMethod Algorithm="http://www.w3.org/2000/09/xmldsig#rsa-sha1"/><ds:Reference URI="#_-4275728805139421773"><ds:Transforms><ds:Transform Algorithm="http://www.w3.org/2000/09/xmldsig#enveloped-signature"/><ds:Transform Algorithm="http://www.w3.org/2001/10/xml-exc-c14n#"/></ds:Transforms><ds:DigestMethod Algorithm="http://www.w3.org/2000/09/xmldsig#sha1"/><ds:DigestValue>m/1fsDUItbWwD3XZbGLhpdIECR0=</ds:DigestValue></ds:Reference></ds:SignedInfo><ds:SignatureValue>pB7HAX8k8bwfrtlvGEKkqWt1WUaw6sZ44tt3Mxn4rxaNhMss2ApnY7VRfbQktDl8EulxugGeSPcuB7odqtlGvOhbNDlNu5U3Ku1fkOZalB7GcwI1/S3AigNLY20LD8GmG6ddHSBJAbo/tLYC/bf+0EuHeKJuPcnGRYR4FDG6H/HS4wj08qb8j3LCFRNqEz6e9wiCKysFd7BSDJsZn/P9QEX/dwnKO13/7u+z7evai4RyfVslMQx4RVYT97f0D9TTSgbYyrSJ2BxkDQak4X+FTLKm85nw1GsTqtzE3pC+WWm2b3soDy9Xvz1DNl9XWdMx0VEswRIQI6IA09bRy3ndOQ==</ds:SignatureValue></ds:Signature><urn:NameIDPolicy Format="urn:oasis:names:tc:SAML:2.0:nameid-format:transient" xmlns:urn="urn:oasis:names:tc:SAML:2.0:protocol"/></urn:AuthnRequest>
2025/04/15 14:17:36 Signature validation failed: Cert is not valid at this time
```
