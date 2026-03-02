import { type ClassValue, clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';

export function cn(...inputs: ClassValue[]) {
	return twMerge(clsx(inputs));
}

export function getInitials(name: string) {
    if (!name) return '';
    const parts = name.split(' ').filter(Boolean);
    if (parts.length === 0) return '';
    if (parts.length === 1) return parts[0].substring(0, 2).toUpperCase();
    return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
}

export function formatCurrency(amount: number) {
    return new Intl.NumberFormat('id-ID', {
        style: 'currency',
        currency: 'IDR',
        minimumFractionDigits: 0,
        maximumFractionDigits: 0
    }).format(amount);
}

export function getStatusBadge(record: any) {
    if (!record) return { label: 'Assigned', class: 'bg-yellow-50 text-yellow-700 border-yellow-200' };

    if (record.paymentStatus === 'Paid') {
        return {
            label: 'Completed',
            class: 'bg-emerald-50 text-emerald-700 border-emerald-200'
        };
    }

    if (record.status === 'Submitted' || record.status === 'Approved') {
        return {
            label: 'In Progress',
            class: 'bg-orange-50 text-orange-700 border-orange-200'
        };
    }

    return {
        label: 'Assigned',
        class: 'bg-yellow-50 text-yellow-700 border-yellow-200'
    };
}
