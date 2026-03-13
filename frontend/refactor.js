import fs from 'fs';
import path from 'path';

const srcDir = path.resolve('src');
const libDir = path.join(srcDir, 'lib');

function ensureDir(dir) {
    if (!fs.existsSync(dir)) {
        fs.mkdirSync(dir, { recursive: true });
    }
}

function moveContents(src, dest) {
    if (fs.existsSync(src)) {
        ensureDir(dest);
        const items = fs.readdirSync(src);
        for (const item of items) {
            const srcPath = path.join(src, item);
            const destPath = path.join(dest, item);
            fs.renameSync(srcPath, destPath);
        }
    }
}

function moveFile(src, dest) {
    if (fs.existsSync(src)) {
        ensureDir(path.dirname(dest));
        fs.renameSync(src, dest);
    }
}

// 1. Create and Move
moveContents(path.join(libDir, 'components/login'), path.join(libDir, 'features/auth/ui'));
moveFile(path.join(libDir, 'stores/auth.ts'), path.join(libDir, 'features/auth/store.ts'));

moveContents(path.join(libDir, 'components/dashboard/admin'), path.join(libDir, 'features/admin/ui'));
moveContents(path.join(libDir, 'components/dashboard/pengajuan'), path.join(libDir, 'features/pengajuan/ui'));
moveFile(path.join(libDir, 'stores/records.ts'), path.join(libDir, 'features/pengajuan/store.ts'));
moveContents(path.join(libDir, 'components/dashboard/laporan'), path.join(libDir, 'features/laporan/ui'));

moveContents(path.join(libDir, 'components/ui'), path.join(libDir, 'shared/ui'));
moveContents(path.join(libDir, 'components/layout'), path.join(libDir, 'shared/ui/layout'));

moveContents(path.join(libDir, 'api'), path.join(libDir, 'shared/api'));

moveFile(path.join(libDir, 'stores/toast.ts'), path.join(libDir, 'shared/stores/toast.ts'));
moveFile(path.join(libDir, 'stores/master-data.ts'), path.join(libDir, 'shared/stores/master-data.ts'));

moveContents(path.join(libDir, 'utils'), path.join(libDir, 'shared/utils'));
moveFile(path.join(libDir, 'utils.ts'), path.join(libDir, 'shared/utils/utils.ts'));

moveContents(path.join(libDir, 'actions'), path.join(libDir, 'shared/actions'));
moveContents(path.join(libDir, 'assets'), path.join(libDir, 'shared/assets'));

const exactReplacements = [
    { from: '$lib/components/login', to: '$lib/features/auth/ui' },
    { from: '$lib/stores/auth', to: '$lib/features/auth/store' },
    { from: '$lib/components/dashboard/admin', to: '$lib/features/admin/ui' },
    { from: '$lib/components/dashboard/pengajuan', to: '$lib/features/pengajuan/ui' },
    { from: '$lib/stores/records', to: '$lib/features/pengajuan/store' },
    { from: '$lib/components/dashboard/laporan', to: '$lib/features/laporan/ui' },
    { from: '$lib/components/ui', to: '$lib/shared/ui' },
    { from: '$lib/components/layout', to: '$lib/shared/ui/layout' },
    { from: '$lib/api', to: '$lib/shared/api' },
    { from: '$lib/stores/toast', to: '$lib/shared/stores/toast' },
    { from: '$lib/stores/master-data', to: '$lib/shared/stores/master-data' },
    { from: '$lib/utils/terbilang', to: '$lib/shared/utils/terbilang' },
    { from: '$lib/utils.ts', to: '$lib/shared/utils/utils.ts' },
    { from: '$lib/utils', to: '$lib/shared/utils/utils' },
    { from: '$lib/actions', to: '$lib/shared/actions' },
    { from: '$lib/assets', to: '$lib/shared/assets' },
];

function processFile(filePath) {
    if (filePath.includes('node_modules') || filePath.includes('.svelte-kit') || filePath.includes('build')) return;
    const stat = fs.statSync(filePath);
    if (stat.isDirectory()) {
        fs.readdirSync(filePath).forEach(file => processFile(path.join(filePath, file)));
    } else if (filePath.endsWith('.svelte') || filePath.endsWith('.ts') || filePath.endsWith('.js')) {
        let content = fs.readFileSync(filePath, 'utf-8');
        let newContent = content;
        
        const sortedReplacements = [...exactReplacements].sort((a, b) => b.from.length - a.from.length);

        for (const { from, to } of sortedReplacements) {
            // escape regex correctly
            const escapedFrom = from.replace(/[-\/\\^$*+?.()|[\]{}]/g, '\\$&');
            const regex = new RegExp(escapedFrom, 'g');
            newContent = newContent.replace(regex, to);
        }

        if (content !== newContent) {
            fs.writeFileSync(filePath, newContent, 'utf-8');
            console.log(`Updated: ${filePath}`);
        }
    }
}

processFile(path.join(srcDir, 'routes'));
processFile(path.join(srcDir, 'lib'));

function cleanEmptyDirs(dir) {
    if (!fs.existsSync(dir)) return;
    const stat = fs.statSync(dir);
    if (stat.isDirectory()) {
        let files = fs.readdirSync(dir);
        files.forEach(file => cleanEmptyDirs(path.join(dir, file)));
        
        files = fs.readdirSync(dir);
        if (files.length === 0) {
            fs.rmdirSync(dir);
            console.log(`Removed empty dir: ${dir}`);
        }
    }
}

cleanEmptyDirs(libDir);
