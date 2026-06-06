# My Odyc game

A small [Odyc.js](https://odyc.dev) game, scaffolded with the Odyc CLI.

## Project structure

```
.
├── index.html      # Loads Odyc + every game file, in order. Open this to play.
├── index.js        # Entry point — starts the first scene.
├── scenes/         # One file per scene (a screen / state of the game).
│   ├── title.js    #   The title screen.
│   └── world.js    #   The playable level.
├── utils/          # Shared, reusable code (the palette, sprites, helpers).
│   └── sprites.js
└── odyc.json       # Links this folder to your game on Odyc (don't edit by hand).
```

There is no build step. Every `.js` file is loaded into one shared scope, so a
function or constant defined in one file is available in the others. To add a
scene, define it as a global function in `scenes/`, add a `<script>` tag for it
in `index.html`, and call it from another scene.

## Run it locally

Serve the folder with any static file server, then open it in your browser:

```bash
npx serve .
# or: python3 -m http.server
```

## Deploy

Publish your changes to Odyc:

```bash
odyc-cli deploy
```

`deploy` bundles `utils/`, `scenes/` and `index.js` (in that order) into a
single file and uploads it. Make sure you're signed in first — a single sign-in
authorizes deploys for all of your games:

```bash
odyc-cli login
```
