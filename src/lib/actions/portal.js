export function portal(node, target = 'body') {
    let targetEl;
    
    async function update() {
        targetEl = document.querySelector(target);
        if (targetEl) {
            targetEl.appendChild(node);
            node.hidden = false;
        } else {
            // Retry if target not found (e.g. hydration mismatch or timing)
            await new Promise(resolve => requestAnimationFrame(resolve));
            targetEl = document.querySelector(target);
            if (targetEl) {
                targetEl.appendChild(node);
                node.hidden = false;
            }
        }
    }

    update();

    return {
        update(newTarget) {
            target = newTarget;
            update();
        },
        destroy() {
            if (node.parentNode) {
                node.parentNode.removeChild(node);
            }
        }
    };
}
