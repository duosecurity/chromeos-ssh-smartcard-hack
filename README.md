# ChromeOS SSH SmartCard Hack

This repository contains some code to duct-tape an SSH agent to a
Chrome extension that implements the [chrome.certificateProvider][]
API. This can make it possible to use Smart Cards with the [Secure Shell][]
app!

This is forked from, and heavily based on, the [MacGyver][] extension
developed by the good folks at Stripe.

## Background

Recently, Chrome OS introduced full [Smart Card support][], along with
the [chrome.certificateProvider][] API that allows Smart Card
middleware extensions to provide certificates (and the ability to sign
data) to ChromeOS for TLS client authentication.

Separately, the [Secure Shell][] extension for Chrome (which is an
OpenSSH compiled for [NaCl][]) [supports][chromium-hterm ssh-agent]
using an external extension as a stand-in for an SSH agent.

The interface offered by [chrome.certificateProvider][] extensions
provides all the necessary cryptographic capabilities to implement a
full-fledged SSH Agent extension that uses keypairs stored on smart
cards. However, at this time, there is no clean way for another
extension to access this functionality. Our solution is, instead, to
inject a modified version of the [MacGyver][] code into a
[chrome.certificateProvider][] extension. It turns out this is
surprisingly straightforward!

## Usage

After installing a properly-hacked extension, you can pass
`--ssh-agent=extensionid` in the "relay options" field (not the "SSH
Arguments"!) of the Secure Shell app.

## Performing the Hack

The SSH Agent code is written in [Go][], and compiled to JavaScript
using [GopherJS][]. Using Go lets us take advantage of packages like
[x/crypto][], which already has an SSH agent implementation.

You can compile the extension by running the following:

 * `go get -u github.com/gopherjs/gopherjs`
 * `cd go && gopherjs build`

Then, take an existing [chrome.certificateProvider][] extension,
unpack it if necessary, and do the following:

 * Copy the 'go' directory and all its contents into the unpacked
   extension directory
 * Edit manifest.json
    * Add `"go/go.js"` as a background script.

      That is, if manifest.json initially contains:

      ```
      "app": {
          "background": {
             "persistent": false,
             "scripts": [ "background.js" ]
          }
      },
      ```

      Then modify it to contain:

      ```
      "app": {
          "background": {
             "persistent": false,
             "scripts": [ "background.js", "go/go.js" ]
          }
      },
      ```

    * Allow messaging from the [Secure Shell][] app.

      That is, add:

      ```
      "externally_connectable": {
          "ids": [
              "pnhechapfaindjhompbnflcldabbghjo",
              "okddffdblfhhnmhodogpojmfkjmhinfp"
              ]
      },
      ```

      as a top-level key. (You may want to use the original
      manifest.json from [MacGyver][] as a reference.)


## Permissions

In order to communicate with smart card hardware, middleware
extensions (typically, those that implement the
[chrome.certificateProvider][] interface) must use Google's [Smart
Card Connector][] app. Unfortunately, this app has a highly
restrictive model for [API permissions][Smart Card Connector API
Permissions], in that it only accepts communications from whitelisted
extensions. When you modify an extension as described above, you will
typically end up with a new extension ID that is not on this
whitelist.

If you are deploying this extension onto enterprise-managed
chromebooks, you can attach policy to the Smart Card Connector app to
override this whitelist, as Google documents in their discussion of
[API permissions][Smart Card Connector API Permissions].

Otherwise, with some JS hackery, it's possible to force-whitelist a
Chrome extension. To do this:

 1. Navigate to chrome://extensions/
 2. Ensure "Developer Mode" is checked
 3. Beneath the "Smart Card Connector" app, click the link to inspect
    the 'background page'
 4. In the JS console, type:
    `$jscomp.scope.permissionsChecker.userPromptingChecker_.storeUserSelection_('YOUR_EXTENSION_ID', true)`

(Obviously, there's a risk this technique might break with a future
update to the Smart Card Connector app!)

## Chrome SSH Agent Protocol

The [Secure Shell][] extension for Chrome has
[supported][chromium-hterm ssh-agent] relaying the SSH agent protocol
to another extension since November 2014. The [protocol][nassh agent]
is fairly straightforward, but undocumented.

The [SSH agent protocol][ssh-agent] is based on a simple
length-prefixed framing protocol. Each message is prefixed with a
4-byte network-encoded length. Messages are sent over a UNIX socket.

By contrast, the [Secure Shell][] agent protocol uses [Chrome
cross-extension messaging][Cross-extension messaging], connecting to
the agent extension with [chrome.runtime.connect][]. Each frame of the
SSH agent protocol is assembled, stripped of its length prefix, and
sent as an array of numbers (not, say, an ArrayBuffer) in the "data"
field of an object via `postMessage`.

Here's an example message, representing the
`SSH2_AGENTC_REQUEST_IDENTITIES` request (to list keys):

```json
{
  "type": "auth-agent@openssh.com",
  "data": [11]
}
```

SSH agents are expected to respond in the same format.

### macgyver.AgentPort

Because [x/crypto][]'s [SSH agent
implementation][x/crypto/ssh/agent.ServeAgent] expects an
[io.ReadWriter][] that implements the standard (length-prefixed)
protocol, MacGyver implements a wrapper around a `chrome.runtime.Port`
that between [Secure Shell][]'s protocol and the native protocol
(stripping or adding the length prefix and JSON object wrapper as
necessary).

## Contributors

MacGyver extension:

* Evan Broder
* Dan Benamy

CertificateProvider modifications:

* Adam Goodman

[Cross-extension messaging]: https://developer.chrome.com/extensions/messaging#external
[Go]: http://golang.org/
[Gopherjs]: http://www.gopherjs.org/
[MacGyver]: https://github.com/stripe/macgyver
[NaCl]: https://en.wikipedia.org/wiki/Google_Native_Client
[Secure Shell]: https://chrome.google.com/webstore/detail/secure-shell/pnhechapfaindjhompbnflcldabbghjo?hl=en
[Smart Card support]: https://support.google.com/chrome/a/answer/7014689?hl=en
[Smart Card Connector]: https://chrome.google.com/webstore/detail/smart-card-connector/khpfeaanjngmcnplbdlpegiifgpfgdco
[Smart Card Connector API Permissions]: https://github.com/GoogleChrome/chromeos_smart_card_connector#smart-card-connector-app-api-permissions
[chrome.certificateProvider]: https://developer.chrome.com/extensions/certificateProvider
[chrome.runtime.connect]: https://developer.chrome.com/extensions/runtime#method-connect
[chromium-hterm ssh-agent]: https://groups.google.com/a/chromium.org/d/msg/chromium-hterm/iq-AuvRJsYw/QVJdCw2wSM0J
[io.ReadWriter]: https://godoc.org/io#ReadWriter
[nassh agent]: https://github.com/libapps/libapps-mirror/blob/master/nassh/js/nassh_stream_sshagent_relay.js
[ssh-agent]: http://cvsweb.openbsd.org/cgi-bin/cvsweb/src/usr.bin/ssh/PROTOCOL.agent?rev=HEAD
[x/crypto/ssh/agent.ServeAgent]: https://godoc.org/golang.org/x/crypto/ssh/agent#ServeAgent
[x/crypto]: https://godoc.org/golang.org/x/crypto
