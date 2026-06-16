<script lang="ts">
	import '../app.css';
	import favicon from '$lib/assets/favicon.svg';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { user } from '$lib/user.svelte';
	import { logout } from '$lib/api';
	import { toaster } from '$lib/toast.svelte';
	import Toaster from '$lib/Toaster.svelte';

	let { children } = $props();

	const PUBLIC = ['/login'];
	const isPublic = $derived(PUBLIC.includes(page.url.pathname));

	onMount(async () => {
		await user.refresh();
		const v = page.url.searchParams.get('verified');
		if (v === '1') toaster.success('Email verified — welcome!');
		else if (v === '0') toaster.error('That verification link is invalid or expired.');
	});

	// Redirect based on auth state once the session check is done.
	$effect(() => {
		if (!user.ready) return;
		if (!user.isAuthed && !isPublic) goto('/login');
		if (user.isAuthed && page.url.pathname === '/login') goto('/');
	});

	async function doLogout() {
		try {
			await logout();
		} catch {
			/* ignore */
		}
		user.set(null);
		goto('/login');
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

{#if user.isAuthed}
	<header>
		<a class="brand" href="/">
			<span class="logo">🥜</span>
			<span class="brand-text">Nut Cracker</span>
		</a>

		<nav>
			<a href="/" class:active={page.url.pathname === '/'}>My Map</a>
			<a href="/friends" class:active={page.url.pathname.startsWith('/friends')}>Friends</a>
			<a href="/settings" class:active={page.url.pathname.startsWith('/settings')}>Settings</a>
			<a href="/profile" class:active={page.url.pathname.startsWith('/profile')}>Profile</a>
		</nav>

		<button class="logout" onclick={doLogout} title="Log out">Log out</button>
	</header>
{/if}

<main class:full={!user.isAuthed}>
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
		gap: 1rem 1.5rem;
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
		flex-wrap: wrap;
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
	.logout {
		margin-left: auto;
		padding: 0.4rem 0.85rem;
		border: 1px solid var(--border);
		border-radius: 999px;
		background: var(--surface);
		color: var(--muted);
		font: inherit;
		font-size: 0.85rem;
		cursor: pointer;
		transition:
			color 0.15s var(--ease),
			border-color 0.15s var(--ease);
	}
	.logout:hover {
		color: var(--danger);
		border-color: color-mix(in srgb, var(--danger) 40%, var(--border));
	}
	main {
		max-width: 1100px;
		margin: 0 auto;
		padding: 1.75rem 1.25rem 4rem;
		animation: fade-in 0.4s var(--ease);
	}
	main.full {
		max-width: none;
		padding: 0;
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

	@media (max-width: 560px) {
		header {
			gap: 0.5rem 0.75rem;
			padding: 0.6rem 0.9rem;
		}
		.logout {
			order: 2;
		}
		nav {
			order: 3;
			width: 100%;
			overflow-x: auto;
			flex-wrap: nowrap;
			scrollbar-width: none;
		}
		nav::-webkit-scrollbar {
			display: none;
		}
		nav a {
			white-space: nowrap;
		}
		main {
			padding: 1.25rem 0.9rem 4rem;
		}
	}
</style>
