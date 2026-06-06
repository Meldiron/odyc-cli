// index.js
//
// Entry point for the game. Both the browser (via index.html) and
// `odyc-cli deploy` load every file in this project into one shared scope —
// utils first, then scenes, then this file last — so by the time this line
// runs, every scene function is defined. We simply start at the title screen.

openTitleScene();
