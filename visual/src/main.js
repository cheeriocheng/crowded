import * as THREE from "three";
import { OrbitControls } from "three/addons/controls/OrbitControls.js";
import { SparkRenderer, SplatMesh } from "@sparkjsdev/spark";

// Clip 2 (Yungang). The numbers are the camera data stored in
// free_range/data/img_gaussian_splat/img.ply: image size and focal length in
// pixels, with the camera at the origin.
const SCENE = {
  url: "/scenes/clip2.splat",
  width: 1696,
  height: 956,
  focal: 1349.92,
};

const frame = document.getElementById("frame");
const status = document.getElementById("status");
const hint = document.getElementById("hint");
frame.style.aspectRatio = `${SCENE.width} / ${SCENE.height}`;

const renderer = new THREE.WebGLRenderer({ antialias: false });
frame.prepend(renderer.domElement);

const scene = new THREE.Scene();
scene.add(new SparkRenderer({ renderer }));

// Same field of view as the photo the splat was made from.
const fov = THREE.MathUtils.radToDeg(2 * Math.atan(SCENE.height / 2 / SCENE.focal));
const camera = new THREE.PerspectiveCamera(fov, SCENE.width / SCENE.height, 0.1, 1000);

// The splat uses the computer-vision convention (y down, z forward). Half a
// turn about x converts it to three.js (y up, camera looking down -z).
const splat = new SplatMesh({ url: SCENE.url });
splat.quaternion.set(1, 0, 0, 0);
scene.add(splat);

splat.initialized.then(
  () => { status.textContent = ""; },
  (err) => {
    console.error(err);
    status.textContent =
      "Could not load the scene. Is visual/public/scenes/clip2.splat in place?";
  },
);

function resize() {
  const { width, height } = frame.getBoundingClientRect();
  renderer.setPixelRatio(Math.min(window.devicePixelRatio || 1, 2));
  renderer.setSize(width, height, false);
}
new ResizeObserver(resize).observe(frame);
resize();

// Inspection: drag to orbit, scroll to zoom, right-drag to pan. The pivot is
// a point in front of the camera at about the building's distance. Press R
// or double-click to return to the original camera position.
const controls = new OrbitControls(camera, renderer.domElement);
controls.target.set(0, 0, -50);
controls.enableDamping = true;
controls.update();
controls.saveState();

const resetView = () => controls.reset();
renderer.domElement.addEventListener("dblclick", resetView);

// Everything below is placed on one plane facing the original camera, at
// this distance. In this scene the eaves are about 3-14 m away, the trees
// about 30 m, and the sky 150 m and beyond.
let depth = 20; // metres
const HALF_W = SCENE.width / 2 / SCENE.focal;
const HALF_H = SCENE.height / 2 / SCENE.focal;
const BALL_SIZE = 0.03; // radius as a share of the distance, so it looks the same size at any depth
const ballShape = new THREE.SphereGeometry(1, 24, 16);

// A flat polygon with one corner pointing up, the same as on the phone.
const polygonShapes = new Map();
function polygonShape(sides) {
  if (!polygonShapes.has(sides)) {
    polygonShapes.set(sides, new THREE.CircleGeometry(1, sides, Math.PI / 2));
  }
  return polygonShapes.get(sides);
}

// Turns a position in the view (-1..1 across, -1..1 up) into a point on the
// plane at the current depth.
function onPlane(nx, ny, out) {
  return out.set(nx * 0.9 * HALF_W * depth, ny * 0.9 * HALF_H * depth, -depth);
}

// Step 1c test object: a bright ball sweeping left and right, to check that
// the scene hides it where the scene is nearer.
const probe = new THREE.Mesh(ballShape, new THREE.MeshBasicMaterial({ color: 0xff2bd6 }));
probe.visible = false;
scene.add(probe);

// Step 1d: one flat polygon per phone, in that phone's colour and shape,
// moved by dragging on the phone. The backend sends every phone's position
// in its shared space.
const balls = new Map(); // id -> { mesh, nx, ny }
let linked = false;

function connect() {
  const scheme = location.protocol === "https:" ? "wss" : "ws";
  const ws = new WebSocket(`${scheme}://${location.host}/ws?role=screen`);
  ws.onopen = () => { linked = true; };
  ws.onmessage = (e) => {
    const m = JSON.parse(e.data);
    if (m.t !== "state") return;
    const seen = new Set();
    for (const p of m.players) {
      seen.add(p.id);
      let b = balls.get(p.id);
      if (!b) {
        const mesh = new THREE.Mesh(polygonShape(p.sides), new THREE.MeshBasicMaterial({ color: p.color }));
        b = { mesh, nx: 0, ny: 0, placed: false };
        balls.set(p.id, b);
        scene.add(mesh);
      }
      b.nx = (p.x / m.w) * 2 - 1;
      b.ny = -((p.y / m.h) * 2 - 1); // the shared space counts y downward
      b.mesh.visible = p.online;
    }
    for (const [id, b] of balls) {
      if (seen.has(id)) continue;
      scene.remove(b.mesh);
      b.mesh.material.dispose();
      balls.delete(id);
    }
  };
  ws.onclose = () => { linked = false; setTimeout(connect, 1000); };
}
connect();

addEventListener("keydown", (e) => {
  if (e.key === "r" || e.key === "R") resetView();
  if (e.key === "t" || e.key === "T") probe.visible = !probe.visible;
  if (e.key === "[") depth = Math.max(2, depth / 1.25);
  if (e.key === "]") depth = Math.min(300, depth * 1.25);
});

let lastHint = "";
function showHint() {
  const online = [...balls.values()].filter((b) => b.mesh.visible).length;
  const phones = linked ? `${online} phone${online === 1 ? "" : "s"}` : "backend not connected";
  const text = `${phones} · shapes at ${depth.toFixed(1)} m · [ nearer · ] farther · T test ball · drag to orbit · R reset`;
  if (text !== lastHint) { hint.textContent = text; lastHint = text; }
}

const target = new THREE.Vector3();
renderer.setAnimationLoop((ms) => {
  if (probe.visible) {
    onPlane(Math.sin(ms / 1000 * 0.6), 0.1, probe.position);
    probe.scale.setScalar(depth * BALL_SIZE);
  }
  for (const b of balls.values()) {
    onPlane(b.nx, b.ny, target);
    if (b.placed) b.mesh.position.lerp(target, 0.25);
    else { b.mesh.position.copy(target); b.placed = true; }
    b.mesh.scale.setScalar(depth * BALL_SIZE);
    b.mesh.quaternion.copy(camera.quaternion); // always face the viewer
  }
  showHint();
  controls.update();
  renderer.render(scene, camera);
});
