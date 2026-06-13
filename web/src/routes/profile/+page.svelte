<script lang="ts">
	import { updateUsername, logout, type Profile } from '$lib/api';
	import { user } from '$lib/user.svelte';
	import { goto } from '$app/navigation';
	import { toaster } from '$lib/toast.svelte';

	let profile = $state<Profile | null>(null);
	let usernameInput = $state('');
	let error = $state('');
	let notice = $state('');
	let saving = $state(false);
	let copied = $state(false);

	$effect(() => {
		const p = user.profile;
		if (p) {
			profile = p;
			usernameInput = p.username;
		}
	});

	async function saveUsername() {
		if (!profile) return;
		notice = '';
		error = '';
		saving = true;
		try {
			const updated = await updateUsername(user.id, usernameInput);
			profile = updated;
			user.set(updated);
			notice = 'Username updated.';
		} catch (e) {
			error = e instanceof Error ? e.message : String(e);
		} finally {
			saving = false;
		}
	}

	async function copyId() {
		if (!profile) return;
		await navigator.clipboard.writeText(String(profile.id));
		copied = true;
		setTimeout(() => (copied = false), 1500);
	}

	async function signOut() {
		try {
			await logout();
		} catch {
			/* ignore */
		}
		user.set(null);
		toaster.success('Signed out.');
		goto('/login');
	}

	const created = $derived(
		profile ? new Date(profile.created_at).toLocaleDateString() : ''
	);
	const dirty = $derived(profile ? usernameInput.trim() !== profile.username : false);
</script>

<h1>Profile</h1>
<p class="muted sub">Your account details. Share your ID so friends can find you.</p>

{#if !user.id}
	<div class="card empty">Enter your ID in the top-right to view your profile.</div>
{:else if !profile}
	{#if error}
		<div class="card empty">⚠️ {error}</div>
	{:else}
		<p class="muted"><span class="spinner"></span> Loading…</p>
	{/if}
{:else}
	<div class="card profile">
		<div class="avatar">{profile.username.slice(0, 2).toUpperCase()}</div>
		<div class="meta">
			<div class="name">{profile.username}</div>
			<div class="muted small">{profile.email}</div>
			<div class="muted small">Joined {created}</div>
		</div>
	</div>

	<div class="card field">
		<div class="field-label">Your ID</div>
		<div class="muted small">Share this number with friends so they can add you.</div>
		<div class="id-row">
			<code class="id">{profile.id}</code>
			<button class="btn btn-ghost" onclick={copyId}>{copied ? 'Copied!' : 'Copy'}</button>
		</div>
	</div>

	<div class="card field">
		<div class="field-label">Username</div>
		<div class="muted small">Your display name, shown to other users.</div>
		<div class="id-row">
			<input
				class="input"
				type="text"
				bind:value={usernameInput}
				maxlength="30"
				onkeydown={(e) => e.key === 'Enter' && dirty && saveUsername()}
			/>
			<button class="btn" disabled={!dirty || saving} onclick={saveUsername}>
				{saving ? 'Saving…' : 'Save'}
			</button>
		</div>
		{#if error}<p class="error">⚠️ {error}</p>{/if}
		{#if notice}<p class="notice">{notice}</p>{/if}
	</div>

	<button class="btn btn-ghost signout" onclick={signOut}>Sign out</button>
{/if}

<style>
	h1 {
		margin: 0;
	}
	.sub {
		margin: 0.25rem 0 1.25rem;
		font-size: 0.92rem;
	}
	.card {
		max-width: 520px;
		margin-bottom: 1rem;
	}
	.profile {
		display: flex;
		align-items: center;
		gap: 1rem;
		padding: 1.1rem 1.25rem;
	}
	.avatar {
		display: grid;
		place-items: center;
		width: 56px;
		height: 56px;
		border-radius: 50%;
		background: linear-gradient(135deg, var(--primary), var(--primary-700));
		color: white;
		font-size: 1.1rem;
		font-weight: 600;
		flex-shrink: 0;
	}
	.meta .name {
		font-size: 1.1rem;
		font-weight: 600;
	}
	.field {
		padding: 1.1rem 1.25rem;
	}
	.field-label {
		font-weight: 600;
		margin-bottom: 0.15rem;
	}
	.id-row {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		margin-top: 0.6rem;
	}
	.id-row .input {
		flex: 1;
	}
	.id {
		flex: 1;
		font-size: 1.3rem;
		font-weight: 700;
		letter-spacing: 0.04em;
		color: var(--primary-700);
		background: var(--surface-2);
		padding: 0.4rem 0.8rem;
		border-radius: var(--radius-sm);
	}
	.small {
		font-size: 0.83rem;
	}
	.empty {
		padding: 1.5rem;
		text-align: center;
		color: var(--muted);
	}
	.signout {
		margin-top: 0.5rem;
	}
	.error {
		color: var(--danger);
		font-size: 0.88rem;
		margin: 0.5rem 0 0;
	}
	.notice {
		color: var(--primary-700);
		font-size: 0.88rem;
		margin: 0.5rem 0 0;
	}
</style>
