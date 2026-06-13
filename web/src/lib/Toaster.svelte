<script lang="ts">
	import { toaster } from '$lib/toast.svelte';
	import { fly, fade } from 'svelte/transition';
	import { flip } from 'svelte/animate';

	const icon = (t: string) => (t === 'success' ? '✓' : t === 'error' ? '⚠' : 'ℹ');
</script>

<div class="toaster" aria-live="polite">
	{#each toaster.items as t (t.id)}
		<button
			class="toast {t.type}"
			animate:flip={{ duration: 200 }}
			in:fly={{ y: 16, duration: 220 }}
			out:fade={{ duration: 150 }}
			onclick={() => toaster.dismiss(t.id)}
		>
			<span class="icon">{icon(t.type)}</span>
			<span class="msg">{t.message}</span>
		</button>
	{/each}
</div>

<style>
	.toaster {
		position: fixed;
		bottom: 1.25rem;
		left: 50%;
		transform: translateX(-50%);
		display: flex;
		flex-direction: column-reverse;
		gap: 0.5rem;
		z-index: 100;
		pointer-events: none;
	}
	.toast {
		pointer-events: auto;
		display: flex;
		align-items: center;
		gap: 0.6rem;
		padding: 0.7rem 1rem;
		border-radius: 12px;
		border: 1px solid var(--border);
		background: var(--surface);
		color: var(--text);
		box-shadow: var(--shadow-lg);
		font: inherit;
		font-size: 0.9rem;
		cursor: pointer;
		min-width: 220px;
		max-width: 90vw;
		text-align: left;
	}
	.icon {
		display: grid;
		place-items: center;
		width: 22px;
		height: 22px;
		border-radius: 50%;
		font-size: 0.8rem;
		font-weight: 700;
		color: white;
		flex-shrink: 0;
	}
	.success .icon {
		background: var(--primary);
	}
	.error .icon {
		background: var(--danger);
	}
	.info .icon {
		background: #3b82f6;
	}
	.error {
		border-color: color-mix(in srgb, var(--danger) 40%, var(--border));
	}
</style>
