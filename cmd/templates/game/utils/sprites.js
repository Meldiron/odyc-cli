// utils/sprites.js
//
// Shared, reusable building blocks for the whole game: the colour palette and
// the sprites every scene draws with. Keeping them here gives you a single
// source of truth — tweak a colour or a sprite once and every scene updates.
//
// In a sprite, each character is an index into PALETTE (0-based) and "." is
// transparent. You can also generate sprites from PNGs with `odyc-cli sprites`.

const BACKGROUND_COLOR = "#1a1c2c";

const PALETTE = [
	"#1a1c2c", // 0 - background / outline
	"#f4f4f4", // 1 - light
	"#41a6f6", // 2 - player
	"#ef7d57", // 3 - walls
	"#a7f070", // 4 - goal
];

const PLAYER_SPRITE = `
	...22...
	...22...
	.222222.
	2.2222.2
	2.2222.2
	..2222..
	..2..2..
	..2..2..
`;

const WALL_SPRITE = `
	33333333
	30303030
	33333333
	03030303
	33333333
	30303030
	33333333
	03030303
`;

const GOAL_SPRITE = `
	.444444.
	4......4
	4.4444.4
	4.4..4.4
	4.4..4.4
	4.4444.4
	4......4
	.444444.
`;
