/** @type {HTMLDivElement} */
let $loader;
/** @type {HTMLDivElement} */
let $pageContainer;
/** @type {HTMLDivElement} */
let $navList;

function renderPages(pages) {
	for (const [i, page] of pages.entries()) {
		const $grid = document.createElement('div');
		$grid.className = 'word-board-grid';
		if (i > 0) {
			$grid.classList.add('page-break');
		}
		for (const tile of page.tiles) {
			const $button = document.createElement('button');
			$button.innerText = tile.label;
			if (typeof tile.handler === 'function') {
				$button.addEventListener('click', () => tile.handler());
			} else {
				$button.addEventListener('click', () => tts(tile.label));
			}
			$grid.appendChild($button);
		}
		$pageContainer.appendChild($grid);
	}
}

function changePage(i) {
	let j = 0;
	for (const child of $pageContainer.children) {
		if (i === j) {
			$navList.children.item(j).ariaCurrent = 'true';
			child.classList.remove('hidden');
		} else {
			$navList.children.item(j).ariaCurrent = 'false';
			child.classList.add('hidden');
		}
		j++;
	}
}

function renderNavigation(pages) {
	const items = [];
	for (const [i, page] of pages.entries()) {
		const $button = document.createElement('button');
		$button.className = 'secondary';
		$button.innerText = page.label;
		$button.addEventListener('click', () => {
			changePage(i);
		});
		items.push($button);
	}
	$navList.replaceChildren(...items);
}

function tts(text) {
	console.log('TTS', text);

	const params = new URLSearchParams();
	params.set('q', text);

	const audio = new Audio(`./tts?${params.toString()}`);
	audio.addEventListener('canplaythrough', () => audio.play(), { once: true });
}

function buttonSound() {
	const audio = new Audio('assets/button1.mp3');
	audio.addEventListener('canplaythrough', () => audio.play(), { once: true });
}

function generateAlphabet() {
	const alphabet = [];

	let i = 'A'.charCodeAt(0);
	let j = 'Z'.charCodeAt(0);
	for (; i <= j; i++) {
		alphabet.push(String.fromCharCode(i));
	}

	alphabet.push(...'ÆØÅ'.split('')); // Add the danish letters

	return alphabet;
}

function generateNumbers() {
	return [...Array(11).keys()];
}

function showLoading(show = true) {
	if (show) {
		$loader.classList.remove('is-hidden');
		return;
	}
	$loader.classList.add('is-hidden');
}

async function getPhrases() {
	showLoading();
	return fetch('/api/phrases')
		.then(resp => {
			if (!resp.ok) {
				return Promise.reject(resp);
			}
			return resp.json();
		})
		.catch(err => {
			console.warn(err);
		})
		.finally(() => {
			showLoading(false);
		});
}

async function getNames() {
	showLoading();
	return fetch('/api/names')
		.then(resp => {
			if (!resp.ok) {
				return Promise.reject(resp);
			}
			return resp.json();
		})
		.catch(err => {
			console.warn(err);
		})
		.finally(() => {
			showLoading(false);
		});
}

export async function main() {
	$loader = document.getElementById('loader');
	$pageContainer = document.querySelector('.page-container');
	$navList = document.getElementById('nav-list');

	const phrases = await getPhrases();
	const names = await getNames();

	const pages = [
		{
			label: 'Fraser',
			tiles: [
				...phrases.map(phrase => ({
					label: phrase,
				})),
			],
		},
		{
			label: 'Navne',
			tiles: [
				...names.map(name => ({
					label: name,
				})),
			],
		},
		{
			label: 'Bogstaver',
			tiles: [
				...generateAlphabet().map(l => ({
					label: l,
					handler: () => {
						wordAccumulator.push(l);
						buttonSound();
					},
				})),
				{
					label: 'Slet',
					handler: () => {
						wordAccumulator = [];
						buttonSound();
					},
				},
				{
					label: 'Næste ord',
					handler: () => {
						tts(wordAccumulator.join(''));
						wordAccumulator = [];
					},
				},
			],
		},
		{
			label: 'Tal',
			tiles: [...generateNumbers().map(n => ({ label: n.toString() }))],
		},
	];

	let wordAccumulator = [];

	renderNavigation(pages);
	renderPages(pages);
	changePage(0);
}
