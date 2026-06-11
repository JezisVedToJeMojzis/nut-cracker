<script lang="ts">
	import favicon from '$lib/assets/favicon.svg';
	import { page } from '$app/state';
	import { user } from '$lib/user.svelte';

	let { children } = $props();
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>

<header>
	<a class="brand" href="/">🥜 Nut Cracker</a>
	<nav>
		<a href="/" class:active={page.url.pathname === '/'}>My Map</a>
		<a href="/friends" class:active={page.url.pathname.startsWith('/friends')}>Friends</a>
	</nav>
	<label class="userid">
		User ID
		<input
			type="text"
			placeholder="paste a user UUID"
			value={user.id}
			oninput={(e) => (user.id = (e.target as HTMLInputElement).value)}
		/>
	</label>
</header>

<main>
	{@render children()}
</main>

<style>
	:global(body) {
		margin: 0;
		font-family: system-ui, sans-serif;
		color: #111827;
		background: #f9fafb;
	}
	header {
		display: flex;
		align-items: center;
		gap: 1.5rem;
		padding: 0.75rem 1.25rem;
		background: white;
		border-bottom: 1px solid #e5e7eb;
		flex-wrap: wrap;
	}
	.brand {
		font-weight: 700;
		font-size: 1.1rem;
		text-decoration: none;
		color: #111827;
	}
	nav {
		display: flex;
		gap: 1rem;
	}
	nav a {
		text-decoration: none;
		color: #6b7280;
		padding: 0.25rem 0;
		border-bottom: 2px solid transparent;
	}
	nav a.active {
		color: #16a34a;
		border-bottom-color: #16a34a;
	}
	.userid {
		margin-left: auto;
		display: flex;
		flex-direction: column;
		font-size: 0.7rem;
		color: #6b7280;
		gap: 0.15rem;
	}
	.userid input {
		padding: 0.3rem 0.5rem;
		border: 1px solid #d1d5db;
		border-radius: 6px;
		min-width: 300px;
		font-size: 0.85rem;
	}
	main {
		max-width: 1100px;
		margin: 0 auto;
		padding: 1.5rem 1rem 3rem;
	}
</style>
