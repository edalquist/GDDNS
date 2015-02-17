## Google Dynamic DNS HTTP API

A little utility [App Engine](https://cloud.google.com/appengine/docs/go/) app written in [Go](https://golang.org) that allows you to configure [Google Domains Dynamic DNS Hosts](https://support.google.com/domains/answer/6147083?hl=en) to be updated via HTTP instead of HTTPS. This is particuarly useful if you want to use the DDNS client built into dd-wrt which does not support HTTPS.

## _**WARNING**_

Using this tool to manage your DDNS records is inherently insecure. Any malicious node between your DDNS client and this GDDNS relay will be able to see and modify your requests. Do not use this to manage important Dynamic DNS records!

## How to Deploy

1. Setup an [App Engine Go development environment](https://cloud.google.com/appengine/docs/go/gettingstarted/devenvironment)
2. Clone this repository
3. Create a [Google Cloud Project](https://console.developers.google.com/)
4. [Deploy the Application](https://cloud.google.com/appengine/docs/go/gettingstarted/uploading)
5. Verify the GDDNS app has deployed correctly by browsing to https://YOUR-APP-ID.appspot.com/admin/domains/list


## How to Configure the GDDNS Relay

1. Setup a [Google Domains Dynamic DNS Host](https://support.google.com/domains/answer/6147083?hl=en)
2. Browse to the GDDNS Admin API: https://YOUR-APP-ID.appspot.com/admin/domains/list
3. Enter the host name, username, and password; click Add Domain
4. The app will generate a big random key to be used as the the domainKey when configuring the DDNS client

