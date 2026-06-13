<script lang="ts">
	import world from '@svg-maps/world';

	type Props = {
		/** Map of UPPERCASE ISO code -> crack count. */
		cracks: Record<string, number>;
		/** Whether to show the numeric count badge / intensity by count. */
		countMode: boolean;
		/** Left-click a country. */
		oncrack: (code: string) => void;
		/** Right-click a country. */
		onuncrack: (code: string) => void;
	};

	let { cracks, countMode, oncrack, onuncrack }: Props = $props();

	type Location = { name: string; id: string; path: string };

	// @svg-maps/world ids are lowercase ISO-2; our backend uses uppercase.
	const locations = (world.locations as Location[]).map((l) => ({
		...l,
		code: l.id.toUpperCase()
	}));

	let hovered = $state<{ name: string; code: string; count: number } | null>(null);

	function fillFor(code: string): string {
		const count = cracks[code] ?? 0;
		if (count <= 0) return '#3a414c';
		if (!countMode) return '#10b981';
		const step = Math.min(count, 6);
		const lightness = 58 - step * 6;
		return `hsl(160, 70%, ${lightness}%)`;
	}

	// ---- Pan & zoom ---------------------------------------------------------
	const MIN_SCALE = 1;
	const MAX_SCALE = 12;

	let scale = $state(1);
	let tx = $state(0);
	let ty = $state(0);

	let viewport = $state<HTMLDivElement | null>(null);
	const pointers = new Map<number, { x: number; y: number }>();
	let lastPinchDist = 0;
	let moved = 0; // total movement during the current gesture (to suppress click)

	function clampScale(s: number) {
		return Math.min(MAX_SCALE, Math.max(MIN_SCALE, s));
	}

	// Keep the map from being panned entirely out of view.
	function clampPan() {
		if (!viewport) return;
		const w = viewport.clientWidth;
		const h = viewport.clientHeight;
		const minX = w - w * scale;
		const minY = h - h * scale;
		tx = Math.min(0, Math.max(minX, tx));
		ty = Math.min(0, Math.max(minY, ty));
	}

	function zoomAt(cx: number, cy: number, factor: number) {
		const next = clampScale(scale * factor);
		if (next === scale) return;
		// Keep the point under (cx, cy) fixed while zooming.
		const worldX = (cx - tx) / scale;
		const worldY = (cy - ty) / scale;
		scale = next;
		tx = cx - worldX * scale;
		ty = cy - worldY * scale;
		clampPan();
	}

	function localPoint(e: { clientX: number; clientY: number }) {
		const r = viewport!.getBoundingClientRect();
		return { x: e.clientX - r.left, y: e.clientY - r.top };
	}

	function onWheel(e: WheelEvent) {
		e.preventDefault();
		const p = localPoint(e);
		zoomAt(p.x, p.y, e.deltaY < 0 ? 1.15 : 1 / 1.15);
	}

	function onPointerDown(e: PointerEvent) {
		pointers.set(e.pointerId, localPoint(e));
		moved = 0;
		if (pointers.size === 2) {
			const [a, b] = [...pointers.values()];
			lastPinchDist = Math.hypot(a.x - b.x, a.y - b.y);
		}
	}

	function onPointerMove(e: PointerEvent) {
		if (!pointers.has(e.pointerId)) return;
		const prev = pointers.get(e.pointerId)!;
		const cur = localPoint(e);
		pointers.set(e.pointerId, cur);

		if (pointers.size === 2) {
			// Pinch zoom around the midpoint.
			const [a, b] = [...pointers.values()];
			const dist = Math.hypot(a.x - b.x, a.y - b.y);
			const midX = (a.x + b.x) / 2;
			const midY = (a.y + b.y) / 2;
			if (lastPinchDist > 0) zoomAt(midX, midY, dist / lastPinchDist);
			lastPinchDist = dist;
			moved += 10;
		} else if (pointers.size === 1) {
			// Pan.
			const dx = cur.x - prev.x;
			const dy = cur.y - prev.y;
			moved += Math.abs(dx) + Math.abs(dy);
			tx += dx;
			ty += dy;
			clampPan();
		}
	}

	function onPointerUp(e: PointerEvent) {
		pointers.delete(e.pointerId);
		if (pointers.size < 2) lastPinchDist = 0;
	}

	function zoomButton(factor: number) {
		if (!viewport) return;
		zoomAt(viewport.clientWidth / 2, viewport.clientHeight / 2, factor);
	}

	function reset() {
		scale = 1;
		tx = 0;
		ty = 0;
	}

	// A click only counts as a crack if the pointer barely moved (not a pan).
	function handleClick(e: MouseEvent, code: string) {
		e.preventDefault();
		if (moved > 6) return;
		oncrack(code);
	}

	function handleContext(e: MouseEvent, code: string) {
		e.preventDefault();
		if (moved > 6) return;
		onuncrack(code);
	}
</script>

<div class="map-wrap">
	<div
		class="viewport"
		bind:this={viewport}
		onwheel={onWheel}
		onpointerdown={onPointerDown}
		onpointermove={onPointerMove}
		onpointerup={onPointerUp}
		onpointercancel={onPointerUp}
		role="application"
		aria-label="Interactive world map"
	>
		<div class="transform" style:transform={`translate(${tx}px, ${ty}px) scale(${scale})`}>
			<svg viewBox={world.viewBox} xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
				{#each locations as loc (loc.code)}
					<!-- svelte-ignore a11y_click_events_have_key_events -->
					<path
						d={loc.path}
						class="country"
						class:cracked={(cracks[loc.code] ?? 0) > 0}
						fill={fillFor(loc.code)}
						role="button"
						tabindex="-1"
						aria-label={loc.name}
						onclick={(e) => handleClick(e, loc.code)}
						oncontextmenu={(e) => handleContext(e, loc.code)}
						onmouseenter={() =>
							(hovered = { name: loc.name, code: loc.code, count: cracks[loc.code] ?? 0 })}
						onmouseleave={() => (hovered = null)}
					/>
				{/each}
			</svg>
		</div>
	</div>

	{#if hovered}
		<div class="tooltip">
			<strong>{hovered.name}</strong>
			{#if hovered.count > 0}
				<span class="count">{hovered.count} crack{hovered.count === 1 ? '' : 's'}</span>
			{/if}
		</div>
	{/if}

	<div class="zoom-controls">
		<button onclick={() => zoomButton(1.4)} aria-label="Zoom in">+</button>
		<button onclick={() => zoomButton(1 / 1.4)} aria-label="Zoom out">−</button>
		<button onclick={reset} aria-label="Reset zoom" class="reset">⤢</button>
	</div>
</div>

<style>
	.map-wrap {
		position: relative;
		width: 100%;
	}
	.viewport {
		position: relative;
		width: 100%;
		overflow: hidden;
		border-radius: var(--radius);
		box-shadow: var(--shadow);
		border: 1px solid var(--border);
		background: linear-gradient(160deg, var(--ocean-1) 0%, var(--ocean-2) 50%, var(--ocean-3) 100%);
		touch-action: none; /* we handle pan/zoom ourselves */
		cursor: grab;
	}
	.viewport:active {
		cursor: grabbing;
	}
	/* Subtle vignette + sheen over the ocean for depth. */
	.viewport::after {
		content: '';
		position: absolute;
		inset: 0;
		pointer-events: none;
		border-radius: var(--radius);
		background:
			radial-gradient(120% 90% at 50% 0%, rgba(255, 255, 255, 0.06), transparent 55%),
			radial-gradient(140% 120% at 50% 110%, rgba(0, 0, 0, 0.45), transparent 60%);
	}
	.transform {
		transform-origin: 0 0;
	}
	svg {
		display: block;
		width: 100%;
		height: auto;
	}
	.country {
		stroke: rgba(20, 24, 30, 0.6);
		stroke-width: 0.4;
		stroke-linejoin: round;
		cursor: pointer;
		transition:
			fill 0.25s var(--ease),
			filter 0.15s var(--ease);
	}
	.country:focus,
	.country:focus-visible {
		outline: none;
	}
	.country.cracked {
		filter: drop-shadow(0 0 1.5px rgba(16, 185, 129, 0.7));
	}
	.country:hover {
		stroke: #f8fafc;
		stroke-width: 0.9;
		filter: brightness(1.15) drop-shadow(0 1px 2px rgba(0, 0, 0, 0.4));
	}
	.tooltip {
		position: absolute;
		top: 12px;
		left: 12px;
		display: flex;
		align-items: center;
		gap: 0.5rem;
		background: rgba(15, 23, 42, 0.9);
		backdrop-filter: blur(6px);
		color: white;
		padding: 0.4rem 0.75rem;
		border-radius: 999px;
		font-size: 0.85rem;
		pointer-events: none;
		box-shadow: var(--shadow-lg);
		animation: tip-in 0.15s var(--ease);
	}
	.count {
		background: rgba(255, 255, 255, 0.18);
		padding: 0.05rem 0.45rem;
		border-radius: 999px;
		font-size: 0.78rem;
		font-weight: 600;
	}
	@keyframes tip-in {
		from {
			opacity: 0;
			transform: translateY(-4px);
		}
		to {
			opacity: 1;
			transform: none;
		}
	}
	.zoom-controls {
		position: absolute;
		right: 12px;
		bottom: 12px;
		display: flex;
		flex-direction: column;
		gap: 0.35rem;
	}
	.zoom-controls button {
		width: 38px;
		height: 38px;
		border: 1px solid var(--border);
		border-radius: 10px;
		background: rgba(37, 41, 47, 0.92);
		backdrop-filter: blur(6px);
		color: var(--text);
		font-size: 1.2rem;
		font-weight: 600;
		cursor: pointer;
		box-shadow: var(--shadow-sm);
		transition:
			background 0.15s var(--ease),
			transform 0.12s var(--ease);
		display: grid;
		place-items: center;
	}
	.zoom-controls button:hover {
		background: var(--surface-2);
		transform: translateY(-1px);
		box-shadow: var(--shadow);
	}
	.zoom-controls .reset {
		font-size: 1rem;
	}
</style>
