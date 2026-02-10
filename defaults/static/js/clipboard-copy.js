/**
 * Clipboard Copy Web Component
 * Based on GitHub's clipboard-copy-element
 * https://github.com/github/clipboard-copy-element
 */

// Clipboard utility functions
function createNode(text) {
  const node = document.createElement('pre');
  node.style.width = '1px';
  node.style.height = '1px';
  node.style.position = 'fixed';
  node.style.top = '5px';
  node.textContent = text;
  return node;
}

function copyNode(node) {
  if ('clipboard' in navigator) {
    return navigator.clipboard.writeText(node.textContent || '');
  }

  const selection = getSelection();
  if (selection == null) {
    return Promise.reject(new Error());
  }

  selection.removeAllRanges();

  const range = document.createRange();
  range.selectNodeContents(node);
  selection.addRange(range);

  document.execCommand('copy');
  selection.removeAllRanges();
  return Promise.resolve();
}

function copyText(text) {
  if ('clipboard' in navigator) {
    return navigator.clipboard.writeText(text);
  }

  const body = document.body;
  if (!body) {
    return Promise.reject(new Error());
  }

  const node = createNode(text);
  body.appendChild(node);
  copyNode(node);
  body.removeChild(node);
  return Promise.resolve();
}

// Show success feedback
function showSuccess(button) {
  const originalLabel = button.getAttribute('aria-label') || 'Copy';
  const feedback = button.getAttribute('data-copy-feedback') || 'Copied!';
  
  // Add success state
  button.classList.add('ClipboardButton--success', 'tooltipped', 'tooltipped-w');
  button.setAttribute('aria-label', feedback);
  
  // Reset after 2 seconds
  setTimeout(() => {
    button.classList.remove('ClipboardButton--success', 'tooltipped', 'tooltipped-w');
    button.setAttribute('aria-label', originalLabel);
  }, 2000);
}

// Copy function
async function copy(button) {
  const id = button.getAttribute('for');
  const text = button.getAttribute('value');

  function trigger() {
    button.dispatchEvent(new CustomEvent('clipboard-copy', {bubbles: true}));
  }

  if (button.getAttribute('aria-disabled') === 'true') {
    return;
  }

  try {
    if (text) {
      await copyText(text);
      showSuccess(button);
      trigger();
    } else if (id) {
      const root = 'getRootNode' in Element.prototype ? button.getRootNode() : button.ownerDocument;
      if (!(root instanceof Document || ('ShadowRoot' in window && root instanceof ShadowRoot))) return;
      const node = root.getElementById(id);
      if (node) {
        await copyTarget(node);
        showSuccess(button);
        trigger();
      }
    }
  } catch (error) {
    // Silently fail
  }
}

function copyTarget(content) {
  if (content instanceof HTMLInputElement || content instanceof HTMLTextAreaElement) {
    return copyText(content.value);
  } else if (content instanceof HTMLAnchorElement && content.hasAttribute('href')) {
    return copyText(content.href);
  } else {
    return copyNode(content);
  }
}

function clicked(event) {
  const button = event.currentTarget;
  if (button instanceof HTMLElement) {
    copy(button);
  }
}

function keydown(event) {
  if (event.key === ' ' || event.key === 'Enter') {
    const button = event.currentTarget;
    if (button instanceof HTMLElement) {
      event.preventDefault();
      copy(button);
    }
  }
}

function focused(event) {
  event.currentTarget.addEventListener('keydown', keydown);
}

function blurred(event) {
  event.currentTarget.removeEventListener('keydown', keydown);
}

// ClipboardCopyElement class
class ClipboardCopyElement extends HTMLElement {
  constructor() {
    super();
    this.addEventListener('click', clicked);
    this.addEventListener('focus', focused);
    this.addEventListener('blur', blurred);
  }

  connectedCallback() {
    if (!this.hasAttribute('tabindex')) {
      this.setAttribute('tabindex', '0');
    }

    if (!this.hasAttribute('role')) {
      this.setAttribute('role', 'button');
    }
  }

  get value() {
    return this.getAttribute('value') || '';
  }

  set value(text) {
    this.setAttribute('value', text);
  }
}

// Define the custom element
if (!customElements.get('clipboard-copy')) {
  customElements.define('clipboard-copy', ClipboardCopyElement);
}

