# crowded — design

## Goal

Run a self-contained interactive visualization on a main screen, driven by
real-time input from many phones in the room, with nothing to install.

People mostly watch the main screen. The phone is a controller: it captures
input and shows only enough to identify the person and guide them.

## Decisions

| Topic | Decision |
|---|---|
| Entry | Visitors join the system's own Wi-Fi; the captive portal popup opens the page. QR codes are the fallback. |
| Phone role | Input device. Shows a colour and polygon matching the person's mark on the main screen, plus brief guidance. |
| Main screen | Driven by the laptop, which is also the development machine. |
| Laptop link | Ethernet cable to the router's LAN port. Carries live data and the laptop's internet. |
| Backend | One small program serving the pages, the WebSocket link and the room logic. Written in Go, built as a single file with the pages embedded. |
| Where the backend runs | On the laptop while developing; on the router for shows, so phones stay connected if the laptop restarts. Same program in both places. |
| Internet | The router has an uplink. The laptop gets internet through it while the visitor network keeps running. The show must still run with the uplink unplugged. |
| Two networks | A captive visitor network (reaches only the backend) and a private network (internet plus backend). |
| Secure pages | Planned: a certificate for a domain we own, renewed automatically whenever the router is online. Needs a domain, and should be decided early because it sets the page address. |
| Scale | 30 concurrent phones to start. Radio, network services and backend stay separable so growth means adding access points and moving the backend, not redesigning. |

## System

```
            internet (optional at show time)
                     |
                  [ WAN ]
        +-------- router ---------+
        |  visitor Wi-Fi (captive)|  ~~~ phones (input, WebSocket)
        |  private network        |
        |  name lookups, addresses|
        |  backend (at shows)     |
        +---------[ LAN ]---------+
                     |  Ethernet
                  laptop
        main-screen visualization
        development; backend (while developing)
                     |
                main screen
```

### Parts

- **Phone page** (`/`): captures touch as relative movement, shows colour and polygon. Must survive
  being closed and reopened, and reconnect by itself.
- **Main-screen page** (`/crowd`): the visualization, opened on the laptop.
- **Backend**: accepts WebSocket connections from phones and displays, keeps
  the list of participants, and sends the combined state to displays at a
  fixed rate. Also answers the phones' connectivity checks with a redirect to
  the phone page, and serves a status page.

### Development workflow

1. **Developing**: backend and both pages run on the laptop. The router sends
   visitor phones to the laptop, so real phones test in-progress code.
2. **Checking on the router**: build for the router's processor, copy the one
   file over, point phones back at the router. Do this regularly.
3. **Show**: backend runs on the router and starts at power-on. The laptop
   only shows the main-screen page.

## Qualities

- **Uplink-optional**: runs without internet at show time; provides internet
  to private devices during development alongside the live visitor network.
- **Low latency**: input should feel live on the main screen.
- **Appliance-like**: power on and it works, including after a power cycle.
- **Cross-platform**: iOS and Android, no app.

## Implementation

Specific products and tools. Kept apart from the system design above, which
does not depend on them.

### Router (not yet chosen)

| | GL.iNet Flint 2 (GL-MT6000) | GL.iNet Beryl AX (GL-MT3000) |
|---|---|---|
| Memory / storage | 1 GB / 8 GB | 512 MB / 256 MB |
| Wired ports | 2 × 2.5G plus 4 × 1G | 1 WAN plus 1 LAN |
| Size and power | Desktop unit, external antennas | Pocket-sized, USB-C power |
| Listed price (Nov 2026) | $159 | about $87 |

Recommended: Flint 2, for its spare wired ports, stronger radio and storage.
The Beryl AX suits a system that has to travel in a bag. Both use the same
processor family, so the software is the same on either.

### Other hardware

- **Ethernet cables**: one for laptop to router, one for wall to router.
- **USB-C to Ethernet adapter**, if the laptop has no Ethernet port.
- **Test phones**: at least one iPhone and one Android.
- **Not hardware, but needed for secure pages**: a domain name.

## Open design work

- **Popup versus web page**: find out what the captive popup can and cannot do
  on each platform (staying open, reconnecting, storage, screen sleep), then
  decide what lives in the popup and what needs the normal browser.
- **Devices that reject no-internet Wi-Fi**: some phones refuse to stay on, or
  silently leave, a network they judge to have no internet. Test the target
  devices and find what keeps them connected.
- **Devices that skip the popup**: phones with a VPN or Android Private DNS
  will not auto-open the page, so the QR path must work by itself.
- **Certificate upkeep**: scheduled renewal on the router, a reminder as a
  backstop, the expiry date shown on the status page, and a pre-show check.


## Discussion

Items raised but not decided. To be talked through later.

- **Portal mechanism**: proposed approach is that the router answers every
  name lookup on the visitor network with the system's own address, and the
  backend redirects the phone's connectivity checks to the page. The
  alternative is openNDS, a ready-made captive portal package for OpenWrt.

## Future ideas
- **Camera and motion input**: available once pages are served securely.
- **Position from a photo**: everyone holds up their phone, and a photo of the
  room establishes where each person is in the space.
- **Colour flash for a camera**: phones flash a given colour so a camera can
  recognise them and work out where people are relative to each other.
- **Find your match**: people look for the person whose phone shows the same
  icon as theirs.

## Terms

| Term | Meaning |
|---|---|
| Router | A small computer with Wi-Fi radios. Hands out addresses, answers name lookups, and passes traffic between its devices and the internet. Ours also runs the backend. |
| WAN port | The router's socket for the outside world, such as an Ethernet cable from the wall. |
| LAN port | The router's socket for a local device, here the laptop. |
| WLAN | The wireless network the router broadcasts. Not the same as WAN. |
| Upstream / uplink | The router's own connection out to the internet. |
| Downstream | Everything that connects to the router: phones and the laptop. |
| Captive portal | The popup a phone shows when it joins a Wi-Fi network that needs a sign-in page, as in hotels. We use it to open our page. |
| Connectivity check | A small web request a phone makes after joining Wi-Fi to see whether the internet is reachable. Redirecting it is what triggers the popup. |
| DNS | The service that turns a name (example.com) into an address. Answering every lookup with our own address is how visitor phones are steered to us. |
| DHCP | The service that gives each device an address when it joins. |
| Backend | The program that receives phone input and sends the combined state to the display. |
| Front end | The pages people see: the phone page and the main-screen page. |
| WebSocket | A connection that stays open so both sides can send messages at any time. |
| HTTP / HTTPS | How pages are fetched. HTTPS is the secure form and needs a certificate. |
| Certificate | Proof to the browser that a page really comes from its domain. Browsers only allow camera and motion sensors on pages that have one. |
| OpenWrt | The Linux system these routers run. |
| Flashing | Replacing the router's whole operating system. |
| Deploying | Copying our program file to the router and restarting it. |

