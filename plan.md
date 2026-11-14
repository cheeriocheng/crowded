# crowded — plan

Each lettered step is small enough to run and review by itself. Work stops
after each one for review. Only steps marked **approved** may be built; the
rest are a proposed order and can change.

Status: `[x]` done · `[~]` written, awaiting review · `[ ]` not started

## 0. Skeleton (done)

- [x] **0a** Backend relays phone input to main screens.
- [x] **0b** Phone page at `/`: colour, polygon, relative drag.
- [x] **0c** Main-screen page at `/crowd`: one mark per phone.

## 1. Clip 2 scene in the web — approved

Uses three.js, Spark and Vite, in the `visual/` folder.

- [x] **1a** Web project runs and shows a framed view that does not fill the
  window.
- [x] **1b** The clip 2 splat appears in the frame, seen from the original
  camera position. The view can be orbited to inspect it; R resets.
- [x] **1c** Check that a test object passing behind the eaves is hidden by
  them. Press T for a test ball; `[` and `]` change its distance.
- [x] **1d** Combine with the skeleton: one shape per phone in the scene, all
  at the same distance, moved by dragging on the phone. Each is a flat
  polygon matching the phone's colour and shape, always facing the viewer.

## 2. Recorded birds

- [ ] **2a** Replay the clip 2 tracks over the scene as flat dots, with
  play and pause.
- [ ] **2b** Fit smooth arcs to the longer track pieces, to get turning
  figures that tracking jitter currently hides.

## 3. Simulated birds

- [ ] **3a** Birds driven by the turning model, flat over the scene, with
  sliders for speed, turn strength, turn persistence and turn bias.
- [ ] **3b** Pull toward the building, and avoidance using the splat's own
  positions. Birds pass behind the eaves.
- [ ] **3c** Adjust the parameters until the simulation resembles the
  recording from 2a and 2b.

## 4. Phones and the control panel

- [ ] **4a** Connect the visual page to the backend so it knows who is
  present.
- [ ] **4b** Touch events: taps, drags and other events sent from phones,
  printed by the backend when started with an option.
- [ ] **4c** Control panel on the same page as the visual view, showing the
  participants and their touch events.
- [ ] **4d** People influence the birds. How is not designed yet.

## 5. More than one experiment

- [ ] **5a** A common shape for an experiment and a way to switch between
  them.
- [ ] **5b** A second experiment: connecting the dots of where people are.

## 6. Sound

- [ ] **6a** Sound input from the laptop available as a signal.
- [ ] **6b** One experiment reacts to it.

## 7. More scenes

- [ ] **7a** Splats for clips 1, 3 and 4.
- [ ] **7b** Several scenes in one space, each in its own frame, with birds
  flying from one to the next.

## 8. Visual polish

- [ ] **8a** Transitions between experiments.

## Ideas

Not planned or ordered yet.

- **Call**: when someone taps or long-presses on their phone, their mark
  gives a deep call like the one in Sky, and ripples spread out from it
  across the screen.
- **White on black**: the white-on-black frame looks really good. Use that
  look for a contrasting section. reference video_birds2_seg.mp4 

## Waiting on hardware

- Router setup, the two networks, and starting at power-on.
- The captive popup, and its behaviour on iPhone and Android.
- Secure pages and certificate upkeep.
- Wi-Fi capacity with real phones, and delay measurement.
