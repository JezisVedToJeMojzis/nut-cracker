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

{#if !user.id}
	<p class="hint">Enter your user ID in the top-right to manage friends.</p>
{:else}
	{#if error}<p class="error">⚠️ {error}</p>{/if}
	{#if notice}<p class="notice">{notice}</p>{/if}

	<section>
		<h2>Add a friend</h2>
		<div class="row">
			<input
				type="text"
				bind:value={targetId}
				placeholder="friend's user UUID"
				onkeydown={(e) => e.key === 'Enter' && send()}
			/>
			<button onclick={send}>Send request</button>
		</div>
	</section>

	<section>
		<h2>Incoming requests ({incoming.length})</h2>
		{#if incoming.length === 0}
			<p class="hint">None.</p>
		{:else}
			<ul>
				{#each incoming as f (f.id)}
					<li>
						<span>{f.username}</span>
						<span class="actions">
							<button onclick={() => act(() => acceptRequest(user.id, f.id), 'Accepted.')}>
								Accept
							</button>
							<button
								class="ghost"
								onclick={() => act(() => declineRequest(user.id, f.id), 'Declined.')}
							>
								Decline
							</button>
						</span>
					</li>
				{/each}
			</ul>
		{/if}
	</section>

	<section>
		<h2>Sent requests ({outgoing.length})</h2>
		{#if outgoing.length === 0}
			<p class="hint">None.</p>
		{:else}
			<ul>
				{#each outgoing as f (f.id)}
					<li>
						<span>{f.username}</span>
						<button class="ghost" onclick={() => act(() => removeFriend(user.id, f.id), 'Cancelled.')}>
							Cancel
						</button>
					</li>
				{/each}
			</ul>
		{/if}
	</section>

	<section>
		<h2>Your friends ({friends.length})</h2>
		{#if friends.length === 0}
			<p class="hint">No friends yet.</p>
		{:else}
			<ul>
				{#each friends as f (f.id)}
					<li>
						<a href={`/friends/${f.id}`}>{f.username}</a>
						<span class="actions">
							<a class="button-link" href={`/friends/${f.id}`}>View map</a>
							<button class="ghost" onclick={() => act(() => removeFriend(user.id, f.id), 'Removed.')}>
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
		margin-top: 0;
	}
	section {
		margin: 1.5rem 0;
	}
	h2 {
		font-size: 1rem;
		margin-bottom: 0.5rem;
	}
	.row {
		display: flex;
		gap: 0.5rem;
	}
	input {
		flex: 1;
		max-width: 360px;
		padding: 0.4rem 0.6rem;
		border: 1px solid #d1d5db;
		border-radius: 6px;
	}
	ul {
		list-style: none;
		padding: 0;
		margin: 0;
		max-width: 480px;
	}
	li {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 0.5rem 0.75rem;
		background: white;
		border: 1px solid #e5e7eb;
		border-radius: 6px;
		margin-bottom: 0.4rem;
	}
	.actions {
		display: flex;
		gap: 0.4rem;
		align-items: center;
	}
	button,
	.button-link {
		padding: 0.3rem 0.7rem;
		border: none;
		background: #16a34a;
		color: white;
		border-radius: 6px;
		cursor: pointer;
		font-size: 0.85rem;
		text-decoration: none;
	}
	button:hover {
		background: #15803d;
	}
	.ghost {
		background: #f3f4f6;
		color: #374151;
	}
	.ghost:hover {
		background: #e5e7eb;
	}
	.hint {
		color: #6b7280;
		font-size: 0.9rem;
	}
	.error {
		color: #b91c1c;
	}
	.notice {
		color: #16a34a;
	}
</style>
