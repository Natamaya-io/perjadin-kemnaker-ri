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

// Cache the NumberFormat instance to prevent severe memory leaks and GC pauses in rendering loops
export const idrFormatter = new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0
});

export function formatCurrency(amount: number) {
    if (typeof amount !== 'number') return idrFormatter.format(0);
    return idrFormatter.format(amount);
}

export function toTitleCase(str: string) {
    if (!str) return '';
    return str.replace(/\w\S*/g, (txt) => {
        const upper = txt.toUpperCase();
        if (upper === 'DKI' || upper === 'DI' || upper === 'NTB' || upper === 'NTT' || upper === 'RI') return upper;
        return txt.charAt(0).toUpperCase() + txt.slice(1).toLowerCase();
    });
}

export function formatLocations(record: any) {
    if (!record) return '-';
    if (record.locations && record.locations.length > 0) {
        return record.locations.map((loc: any) => `${toTitleCase(loc.location)}, ${toTitleCase(loc.province)}`).join(' & ');
    }
    return `${toTitleCase(record.location)}, ${toTitleCase(record.province)}`;
}

export function getStatusBadge(record: any) {
    if (!record) return { label: 'Assigned', class: 'bg-yellow-50 text-yellow-700 border-yellow-200' };

    // Group evaluation (if the record has employeesList, it represents the whole SPD group)
    if (record.employeesList && Array.isArray(record.employeesList) && record.employeesList.length > 0) {
        // Only Completed if ALL employees are Paid
        const allPaid = record.employeesList.every((emp: any) => emp.paymentStatus === 'Paid');
        if (allPaid) {
            return {
                label: 'Completed',
                class: 'bg-emerald-50 text-emerald-700 border-emerald-200'
            };
        }

        // Check if any is Rejected
        const anyRejected = record.employeesList.some((emp: any) => emp.status === 'Rejected');
        if (anyRejected) {
             return {
                 label: 'Rejected',
                 class: 'bg-red-50 text-red-700 border-red-200'
             };
        }

        // Otherwise it's In Progress if any is Submitted/Approved or Paid (but not all)
        const anyInProgress = record.employeesList.some((emp: any) => emp.status === 'Submitted' || emp.status === 'Approved' || emp.paymentStatus === 'Paid');
        if (anyInProgress) {
            return {
                label: 'In Progress',
                class: 'bg-orange-50 text-orange-700 border-orange-200'
            };
        }

        // Default to Assigned
        return { label: 'Assigned', class: 'bg-yellow-50 text-yellow-700 border-yellow-200' };
    }

    // Individual evaluation
    if (record.paymentStatus === 'Paid') {
        return {
            label: 'Completed',
            class: 'bg-emerald-50 text-emerald-700 border-emerald-200'
        };
    }

    if (record.status === 'Rejected') {
        return {
            label: 'Rejected',
            class: 'bg-red-50 text-red-700 border-red-200'
        };
    }

    if (record.status === 'Submitted' || record.status === 'Approved') {
        return {
            label: 'In Progress',
            class: 'bg-orange-50 text-orange-700 border-orange-200'
        };
    }

    return { label: 'Assigned', class: 'bg-yellow-50 text-yellow-700 border-yellow-200' };
}
