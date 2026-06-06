// scenes/world.js
//
// The main, playable scene. The player walks a small maze to reach the goal.
// Walls are solid; stepping on the goal ends the game with a message. The
// sprites and palette come from utils/sprites.js.

function openWorldScene() {
	return createGame({
		title: "MY ODYC GAME",
		background: BACKGROUND_COLOR,
		colors: PALETTE,
		player: {
			sprite: PLAYER_SPRITE,
			position: [1, 1],
		},
		templates: {
			"#": {
				sprite: WALL_SPRITE,
				solid: true,
			},
			"*": {
				sprite: GOAL_SPRITE,
				end: "You reached the goal!",
			},
		},
		map: `
			########
			#......#
			#.####.#
			#....#.#
			#.##.#.#
			#..#...#
			#.####*#
			########
		`,
	});
}
