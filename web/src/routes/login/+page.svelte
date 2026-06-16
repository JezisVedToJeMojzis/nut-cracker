<script lang="ts">
	import { goto } from '$app/navigation';
	import { login, register } from '$lib/api';
	import { user } from '$lib/user.svelte';
	import { toaster } from '$lib/toast.svelte';

	type Mode = 'login' | 'register';
	let mode = $state<Mode>('login');

	let email = $state('');
	let username = $state('');
	let password = $state('');
	let busy = $state(false);

	async function submit(e: SubmitEvent) {
		e.preventDefault();
		if (busy) return;
		busy = true;
		try {
			if (mode === 'login') {
				user.set(await login(email, password));
			} else {
				user.set(await register(email, username, password));
				toaster.success('Welcome to Nut Cracker! 🥜');
			}
			goto('/');
		} catch (err) {
			toaster.error(err instanceof Error ? err.message : String(err));
		} finally {
			busy = false;
		}
	}
</script>

<div class="wrap">
	<div class="card">
		<div class="brand"><span class="logo">🥜</span> Nut Cracker</div>

		<div class="tabs">
			<button class:active={mode === 'login'} onclick={() => (mode = 'login')}>Log in</button>
			<button class:active={mode === 'register'} onclick={() => (mode = 'register')}>Sign up</button>
		</div>

		<form onsubmit={submit}>
			<label>
				Email
				<input class="input" type="email" bind:value={email} required autocomplete="email" />
			</label>

			{#if mode === 'register'}
				<label>
					Username
					<input
						class="input"
						type="text"
						bind:value={username}
						required
						minlength="2"
						maxlength="30"
					/>
				</label>
			{/if}

			<label>
				Password
				<input
					class="input"
					type="password"
					bind:value={password}
					required
					minlength="8"
					autocomplete={mode === 'login' ? 'current-password' : 'new-password'}
				/>
			</label>

			<button class="btn submit" type="submit" disabled={busy}>
				{#if busy}<span class="spinner"></span>{/if}
				{mode === 'login' ? 'Log in' : 'Create account'}
			</button>
		</form>
	</div>
</div>

<style>
	.wrap {
		min-height: 100vh;
		display: grid;
		place-items: center;
		padding: 1.5rem;
	}
	.card {
		width: 100%;
		max-width: 380px;
		background: var(--surface);
		border: 1px solid var(--border);
		border-radius: var(--radius);
		box-shadow: var(--shadow-lg);
		padding: 1.75rem;
	}
	.brand {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		font-weight: 700;
		font-size: 1.2rem;
		justify-content: center;
		margin-bottom: 1.25rem;
	}
	.logo {
		font-size: 1.5rem;
	}
	.tabs {
		display: flex;
		gap: 0.25rem;
		background: var(--surface-2);
		padding: 0.25rem;
		border-radius: 10px;
		margin-bottom: 1.25rem;
	}
	.tabs button {
		flex: 1;
		padding: 0.5rem;
		border: none;
		background: transparent;
		color: var(--muted);
		font: inherit;
		font-weight: 500;
		border-radius: 8px;
		cursor: pointer;
	}
	.tabs button.active {
		background: var(--surface);
		color: var(--text);
		box-shadow: var(--shadow-sm);
	}
	form {
		display: flex;
		flex-direction: column;
		gap: 0.9rem;
	}
	label {
		display: flex;
		flex-direction: column;
		gap: 0.3rem;
		font-size: 0.85rem;
		color: var(--muted);
	}
	.submit {
		margin-top: 0.25rem;
		padding: 0.6rem;
		font-size: 0.95rem;
	}
</style>
