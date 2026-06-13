<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { resetPassword } from '$lib/api';
	import { toaster } from '$lib/toast.svelte';

	const token = $derived(page.url.searchParams.get('token') ?? '');
	let password = $state('');
	let busy = $state(false);

	async function submit(e: SubmitEvent) {
		e.preventDefault();
		if (busy || !token) return;
		busy = true;
		try {
			await resetPassword(token, password);
			toaster.success('Password updated — you can log in now.');
			goto('/login');
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
		<h2>Choose a new password</h2>

		{#if !token}
			<p class="muted">This reset link is missing its token. Request a new one from the login page.</p>
			<a class="btn" href="/login">Back to log in</a>
		{:else}
			<form onsubmit={submit}>
				<label>
					New password
					<input
						class="input"
						type="password"
						bind:value={password}
						required
						minlength="8"
						autocomplete="new-password"
					/>
				</label>
				<button class="btn submit" type="submit" disabled={busy}>
					{#if busy}<span class="spinner"></span>{/if}
					Update password
				</button>
			</form>
		{/if}
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
		margin-bottom: 1rem;
	}
	.logo {
		font-size: 1.5rem;
	}
	h2 {
		text-align: center;
		margin: 0 0 1.25rem;
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
		padding: 0.6rem;
	}
</style>
