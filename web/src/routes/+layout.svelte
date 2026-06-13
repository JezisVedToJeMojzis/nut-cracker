<script lang="ts">
	import '../app.css';
	import favicon from '$lib/assets/favicon.svg';
	import { page } from '$app/state';
	import { user } from '$lib/user.svelte';
	import Toaster from '$lib/Toaster.svelte';

	let { children } = $props();

	let signInValue = $state('');
	function signIn() {
		const v = signInValue.trim();
		if (v) user.id = v;
	}
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
	<link rel="preconnect" href="https://fonts.googleapis.com" />
	<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin="anonymous" />
	<link
		href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap"
		rel="stylesheet"
	/>
</svelte:head>

<header>
	<a class="brand" href="/">
		<span class="logo">🥜</span>
		<span>Nut Cracker</span>
	</a>

	<nav>
		<a href="/" class:active={page.url.pathname === '/'}>My Map</a>
		<a href="/friends" class:active={page.url.pathname.startsWith('/friends')}>Friends</a>
		<a href="/settings" class:active={page.url.pathname.startsWith('/settings')}>Settings</a>
		<a href="/profile" class:active={page.url.pathname.startsWith('/profile')}>Profile</a>
	</nav>

	{#if !user.id}
		<div class="user">
			<input
				class="input"
				type="text"
				inputmode="numeric"
				placeholder="enter your ID"
				bind:value={signInValue}
				onkeydown={(e) => e.key === 'Enter' && signIn()}
			/>
		</div>
	{/if}
</header>

<main>
	{@render children()}
</main>

<Toaster />

<style>
	header {
		position: sticky;
		top: 0;
		z-index: 10;
		display: flex;
		align-items: center;
		gap: 1.5rem;
		padding: 0.7rem 1.25rem;
		background: var(--header-bg);
		backdrop-filter: saturate(180%) blur(12px);
		border-bottom: 1px solid var(--border);
		flex-wrap: wrap;
	}
	.brand {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		font-weight: 700;
		font-size: 1.05rem;
		letter-spacing: -0.02em;
		text-decoration: none;
		color: var(--text);
	}
	.logo {
		font-size: 1.3rem;
		filter: drop-shadow(0 2px 3px rgba(0, 0, 0, 0.15));
	}
	nav {
		display: flex;
		gap: 0.35rem;
	}
	nav a {
		text-decoration: none;
		color: var(--muted);
		padding: 0.35rem 0.8rem;
		border-radius: 999px;
		font-size: 0.9rem;
		font-weight: 500;
		transition:
			background 0.15s var(--ease),
			color 0.15s var(--ease);
	}
	nav a:hover {
		background: var(--surface-2);
		color: var(--text);
	}
	nav a.active {
		background: var(--primary-soft);
		color: var(--primary-700);
	}
	.user {
		margin-left: auto;
	}
	.user .input {
		min-width: 220px;
		font-size: 0.82rem;
	}
	main {
		max-width: 1100px;
		margin: 0 auto;
		padding: 1.75rem 1.25rem 4rem;
		animation: fade-in 0.4s var(--ease);
	}
	@keyframes fade-in {
		from {
			opacity: 0;
			transform: translateY(6px);
		}
		to {
			opacity: 1;
			transform: none;
		}
	}
</style>
