_Please note that this project is currently a work in progress. Not everything
will fully work yet_

# Nuts demo EHR system

This application is pretending to be an electronic health record system. You can
use it to demo how healthcare professionals can work together by sharing
information with colleagues through the Nuts nodes.

Also, you can use it as "the other party" in implementing Nuts in your own EHR
systems.

**NOTE THAT THIS APPLICATION IS NOT INTENDED FOR USE WITH REAL MEDICAL
INFORMATION! IT IS IN NO WAY DEVELOPED TO BE SAFE, STABLE OR EVEN USABLE FOR
SUCH PURPOSE.**

## Quick start guide

```bash
git clone git@github.com:nuts-foundation/nuts-demo-ehr.git
cd nuts-demo-ehr
npm install
npm start
```

You should now have three instances of this EHR running on:

* http://localhost:80 ⸺ Verpleeghuis de Nootjes
* http://localhost:81 ⸺ Huisartsenpraktijk Nootenboom
* http://localhost:82 ⸺ Medisch Centrum Noot aan de Man

Also, as a bonus, you can display two or all three side by side by going to:

* http://localhost/duo.html ⸺ Shows the applications on ports 80 and 81
* http://localhost/triple.html ⸺ Shows all three applications

### Configuring the application(s)

You can find the configuration files for all three applications in the `config`
directory. You may need to edit these files to point to the right Nuts node(s).
If you followed the [Setup a local Nuts network](https://nuts-documentation.readthedocs.io/en/latest/pages/getting_started/local_network.html#setup-a-local-nuts-network)
instructions, http://localhost:80 (Verpleeghuis de Nootjes) will connect to Nuts
node 'Bundy' and the other two will connect to node 'Dahmer'.

You can also change port numbers, organisation details and default health
records in the config files.

### Adding to the Nuts register

If you want to allow the applications to find each other and exchange data, you
will have to add them to the Nuts registry. Again, if you followed
[Setup a local Nuts network](https://nuts-documentation.readthedocs.io/en/latest/pages/getting_started/local_network.html#setup-a-local-nuts-network),
you can find your registry in `nuts-network-local/config/registry`.

#### 1. Add the organisations

To make this process easier the applications will output their registry
information on startup. Add that information to your registry's
`organisations.json`.

#### 2. Add the endpoints

You can add the locations of the APIs to the `endpoints.json` file as endpoints
of the type `urn:ietf:rfc:3986:urn:oid:1.3.6.1.4.1.54851.2:demo-ehr`. Also, each
Nuts node that can receive consent needs an endpoint of the type
`urn:nuts:endpoint:consent`.

So for each application add this endpoint to `endpoints.json`:

```json
{
  "endpointType": "urn:ietf:rfc:3986:urn:oid:1.3.6.1.4.1.54851.2:demo-ehr",
  "identifier": "0e906b06-db48-452c-bb61-559f239a06ca",
  "status": "active",
  "version": "0.1.0",
  "URL": "http://localhost:80"
}
```

Make sure you give each one a unique identifier and have it point to the right
URL. Also, make sure both Nuts nodes have a consent endpoint (should be okay if
you're using `nuts-network-local`).

#### 3. Connect the endpoints to organisations

Connect your endpoints to organizations in the `endpoint_organizations.json`
file like this:

```json
{
  "status": "active",
  "organization": "urn:oid:2.16.840.1.113883.2.4.6.1:12345678",
  "endpoint": "0e906b06-db48-452c-bb61-559f239a06ca"
}
```

Make sure you add two entries for each organisation, one for the API and one for
the Nuts node consent endpoint.

Note that for hot reloading of the registry to be triggered, every one of these
three file needs to be touched.

## Learning from this application

If you're curious as to how this application interfaces with the Nuts node,
please take a look at [`resources/nuts-node`](resources/nuts-node), where we
define the different services and API calls that the Nuts node exposes. For
examples on how we then use those services, you can check out the [client APIs](client-api)
that the browser talks to to get things done. Mainly [`consent.js`](client-api/consent.js)
and [`organisation.js`](client-api/organisation.js). Also, we register our
applications on the Nuts node in the root [`index.js`](index.js).
