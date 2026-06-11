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

	// Fill color: uncracked = light grey; cracked = green, deeper with more
	// cracks when count mode is on.
	function fillFor(code: string): string {
		const count = cracks[code] ?? 0;
		if (count <= 0) return '#e5e7eb';
		if (!countMode) return '#16a34a';
		// Count mode: ramp from light to dark green over counts 1..6+.
		const step = Math.min(count, 6);
		const lightness = 60 - step * 6; // 54% .. 24%
		return `hsl(142, 60%, ${lightness}%)`;
	}

	function handleClick(e: MouseEvent, code: string) {
		e.preventDefault();
		oncrack(code);
	}

	function handleContext(e: MouseEvent, code: string) {
		e.preventDefault();
		onuncrack(code);
	}
</script>

<div class="map-wrap">
	<svg viewBox={world.viewBox} xmlns="http://www.w3.org/2000/svg" role="group" aria-label="World map">
		{#each locations as loc (loc.code)}
			<!-- svelte-ignore a11y_click_events_have_key_events -->
			<path
				d={loc.path}
				class="country"
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

	{#if hovered}
		<div class="tooltip">
			<strong>{hovered.name}</strong>
			{#if hovered.count > 0}
				<span class="count">· {hovered.count} crack{hovered.count === 1 ? '' : 's'}</span>
			{/if}
		</div>
	{/if}
</div>

<style>
	.map-wrap {
		position: relative;
		width: 100%;
	}
	svg {
		width: 100%;
		height: auto;
		background: #cfe8ff;
		border-radius: 8px;
	}
	.country {
		stroke: #ffffff;
		stroke-width: 0.5;
		cursor: pointer;
		transition: fill 0.1s ease;
	}
	.country:hover {
		stroke: #111827;
		stroke-width: 1;
	}
	.tooltip {
		position: absolute;
		top: 8px;
		left: 8px;
		background: rgba(17, 24, 39, 0.85);
		color: white;
		padding: 4px 10px;
		border-radius: 6px;
		font-size: 14px;
		pointer-events: none;
	}
	.count {
		opacity: 0.85;
	}
</style>
