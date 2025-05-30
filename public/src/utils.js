/**
 * @param {function} onReady
 */
export function onDocumentReady(onReady) {
	if (typeof onReady !== 'function') {
		throw new Error('onReady argument must be a function');
	}

	if (
		document.readyState === 'complete' ||
		(document.readyState !== 'loading' && !document.documentElement.doScroll)
	) {
		onReady();
	} else {
		document.addEventListener('DOMContentLoaded', onReady);
	}
}
