<script lang="ts">
	import {
		listFriends,
		listIncoming,
		listOutgoing,
		sendRequest,
		acceptRequest,
		declineRequest,
		removeFriend,
		type Friend
	} from '$lib/api';
	import { user } from '$lib/user.svelte';
	import { fade, slide } from 'svelte/transition';

	const initials = (name: string) => name.slice(0, 2).toUpperCase();

	let friends = $state<Friend[]>([]);
	let incoming = $state<Friend[]>([]);
	let outgoing = $state<Friend[]>([]);
	let targetId = $state('');
	let error = $state('');
	let notice = $state('');

	$effect(() => {
		if (user.id) refresh(user.id);
	});

	async function refresh(id: string) {
		error = '';
		try {
			[friends, incoming, outgoing] = await Promise.all([
				listFriends(id),
				listIncoming(id),
				listOutgoing(id)
			]);
		} catch (e) {
			error = e instanceof Error ? e.message : String(e);
		}
	}

	async function act(fn: () => Promise<unknown>, ok: string) {
		error = '';
		notice = '';
		try {
			await fn();
			notice = ok;
			await refresh(user.id);
		} catch (e) {
			error = e instanceof Error ? e.message : String(e);
		}
	}

	function send() {
		if (!targetId.trim()) return;
		const to = targetId.trim();
		act(() => sendRequest(user.id, to), 'Request sent.').then(() => (targetId = ''));
	}
</script>

<h1>Friends</h1>
<p class="muted sub">Add friends to share your map with them.</p>

{#if !user.id}
	<div class="card empty">Enter your user ID in the top-right to manage friends.</div>
{:else}
	{#if error}<p class="banner error" transition:slide>⚠️ {error}</p>{/if}
	{#if notice}<p class="banner notice" transition:slide>{notice}</p>{/if}

	<section class="card">
		<h2>Add a friend</h2>
		<div class="row">
			<input
				class="input"
				type="text"
				bind:value={targetId}
				placeholder="friend's user UUID"
				onkeydown={(e) => e.key === 'Enter' && send()}
			/>
			<button class="btn" onclick={send}>Send request</button>
		</div>
	</section>

	{#if incoming.length > 0}
		<section>
			<h2>Incoming requests <span class="badge">{incoming.length}</span></h2>
			<ul>
				{#each incoming as f (f.id)}
					<li class="card" transition:slide>
						<span class="who">
							<span class="avatar">{initials(f.username)}</span>
							{f.username}
						</span>
						<span class="actions">
							<button class="btn" onclick={() => act(() => acceptRequest(user.id, f.id), 'Accepted.')}>
								Accept
							</button>
							<button
								class="btn btn-ghost"
								onclick={() => act(() => declineRequest(user.id, f.id), 'Declined.')}
							>
								Decline
							</button>
						</span>
					</li>
				{/each}
			</ul>
		</section>
	{/if}

	{#if outgoing.length > 0}
		<section>
			<h2>Sent requests <span class="badge">{outgoing.length}</span></h2>
			<ul>
				{#each outgoing as f (f.id)}
					<li class="card" transition:slide>
						<span class="who">
							<span class="avatar pending">{initials(f.username)}</span>
							{f.username}
							<span class="muted small">pending…</span>
						</span>
						<button
							class="btn btn-ghost"
							onclick={() => act(() => removeFriend(user.id, f.id), 'Cancelled.')}
						>
							Cancel
						</button>
					</li>
				{/each}
			</ul>
		</section>
	{/if}

	<section>
		<h2>Your friends <span class="badge">{friends.length}</span></h2>
		{#if friends.length === 0}
			<div class="card empty">No friends yet — send a request above.</div>
		{:else}
			<ul>
				{#each friends as f (f.id)}
					<li class="card" transition:slide>
						<a class="who" href={`/friends/${f.id}`}>
							<span class="avatar">{initials(f.username)}</span>
							{f.username}
						</a>
						<span class="actions">
							<a class="btn" href={`/friends/${f.id}`}>View map</a>
							<button
								class="btn btn-danger"
								onclick={() => act(() => removeFriend(user.id, f.id), 'Removed.')}
							>
								Unfriend
							</button>
						</span>
					</li>
				{/each}
			</ul>
		{/if}
	</section>
{/if}

<style>
	h1 {
		margin: 0;
	}
	.sub {
		margin: 0.25rem 0 1.25rem;
		font-size: 0.92rem;
	}
	section {
		margin: 1.25rem 0;
	}
	section.card {
		padding: 1.1rem 1.25rem;
	}
	h2 {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		margin: 0 0 0.75rem;
	}
	.row {
		display: flex;
		gap: 0.5rem;
	}
	.row .input {
		flex: 1;
		max-width: 380px;
	}
	ul {
		list-style: none;
		padding: 0;
		margin: 0;
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
		max-width: 560px;
	}
	li {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 0.6rem 0.85rem;
		transition:
			box-shadow 0.15s var(--ease),
			transform 0.15s var(--ease);
	}
	li:hover {
		box-shadow: var(--shadow);
		transform: translateY(-1px);
	}
	.who {
		display: flex;
		align-items: center;
		gap: 0.6rem;
		font-weight: 550;
		color: var(--text);
		text-decoration: none;
	}
	.avatar {
		display: grid;
		place-items: center;
		width: 34px;
		height: 34px;
		border-radius: 50%;
		background: linear-gradient(135deg, var(--primary), var(--primary-700));
		color: white;
		font-size: 0.78rem;
		font-weight: 600;
		flex-shrink: 0;
	}
	.avatar.pending {
		background: linear-gradient(135deg, #cbd5e1, #94a3b8);
	}
	.small {
		font-size: 0.78rem;
		font-weight: 400;
	}
	.actions {
		display: flex;
		gap: 0.4rem;
		align-items: center;
	}
	.empty {
		padding: 1.5rem;
		text-align: center;
		color: var(--muted);
	}
	.banner {
		padding: 0.6rem 0.9rem;
		border-radius: var(--radius-sm);
		font-size: 0.9rem;
		margin: 0.5rem 0;
	}
	.banner.error {
		background: var(--danger-soft);
		color: var(--danger);
	}
	.banner.notice {
		background: var(--primary-soft);
		color: var(--primary-700);
	}
</style>
