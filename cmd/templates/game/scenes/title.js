// scenes/title.js
//
// The title screen. It shows a short intro message and then hands control to
// the world scene. Each scene owns its own `createGame` call, which keeps
// scenes independent and easy to reason about.

async function openTitleScene() {
	const scene = createGame({
		background: BACKGROUND_COLOR,
		colors: PALETTE,
	});

	await scene.openMessage(
		"MY ODYC GAME\n\nReach the glowing goal!\nMove with the arrow keys or WASD.",
	);

	openWorldScene();
}
