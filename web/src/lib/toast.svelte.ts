// A tiny global toast system. Call toaster.success/error/info from anywhere;
// the <Toaster /> component (mounted in the layout) renders them.

export type ToastType = 'success' | 'error' | 'info';
export type Toast = { id: number; type: ToastType; message: string };

let items = $state<Toast[]>([]);
let nextId = 0;

function push(type: ToastType, message: string, ttl = 3500) {
	const id = ++nextId;
	items.push({ id, type, message });
	setTimeout(() => dismiss(id), ttl);
}

function dismiss(id: number) {
	items = items.filter((t) => t.id !== id);
}

export const toaster = {
	get items(): Toast[] {
		return items;
	},
	success: (m: string) => push('success', m),
	error: (m: string) => push('error', m),
	info: (m: string) => push('info', m),
	dismiss
};
