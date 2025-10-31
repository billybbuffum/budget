// Global state
let currentMonth = new Date();
let accounts = [];
let categories = [];
let categoryGroups = [];
let transactions = [];
let allocations = [];
let collapsedGroups = new Set(); // Track which groups are collapsed

// Theme management
function initializeTheme() {
    const savedTheme = localStorage.getItem('theme');
    const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
    const theme = savedTheme || (prefersDark ? 'dark' : 'light');

    if (theme === 'dark') {
        document.documentElement.classList.add('dark');
        updateThemeIcon('dark');
    } else {
        document.documentElement.classList.remove('dark');
        updateThemeIcon('light');
    }
}

function toggleTheme() {
    const isDark = document.documentElement.classList.toggle('dark');
    const theme = isDark ? 'dark' : 'light';
    localStorage.setItem('theme', theme);
    updateThemeIcon(theme);
}

function updateThemeIcon(theme) {
    const icon = document.getElementById('theme-icon');
    if (icon) {
        icon.textContent = theme === 'dark' ? '☀️' : '🌙';
    }
}

// Make toggleTheme available globally for onclick handler
window.toggleTheme = toggleTheme;

// Utility functions
function formatCurrency(cents) {
    return new Intl.NumberFormat('en-US', {
        style: 'currency',
        currency: 'USD'
    }).format(cents / 100);
}

function formatDate(dateString) {
    return new Date(dateString).toLocaleDateString('en-US', {
        month: 'short',
        day: 'numeric',
        year: 'numeric'
    });
}

function getCurrentPeriod() {
    const year = currentMonth.getFullYear();
    const month = String(currentMonth.getMonth() + 1).padStart(2, '0');
    return `${year}-${month}`;
}

function formatMonthYear() {
    return currentMonth.toLocaleDateString('en-US', {
        month: 'long',
        year: 'numeric'
    });
}

// Show toast notification
function showToast(message, type = 'success') {
    const toast = document.getElementById('toast');
    const toastMessage = document.getElementById('toast-message');

    toastMessage.textContent = message;
    toast.className = 'toast active ' + (type === 'error' ? 'bg-red-600' : 'bg-green-600');

    setTimeout(() => {
        toast.className = 'toast';
    }, 3000);
}

// API functions
async function apiCall(endpoint, options = {}) {
    try {
        const response = await fetch(`/api${endpoint}`, {
            headers: {
                'Content-Type': 'application/json',
                ...options.headers
            },
            ...options
        });

        if (!response.ok) {
            const error = await response.text();
            throw new Error(error || `HTTP ${response.status}`);
        }

        // Check if response has content
        const text = await response.text();
        return text ? JSON.parse(text) : null;
    } catch (error) {
        console.error('API call failed:', error);
        showToast(error.message, 'error');
        throw error;
    }
}

// Load data functions
async function loadAccounts() {
    accounts = await apiCall('/accounts') || [];
    return accounts;
}

async function loadCategories() {
    categories = await apiCall('/categories') || [];
    return categories;
}

async function loadCategoryGroups() {
    categoryGroups = await apiCall('/category-groups') || [];
    return categoryGroups;
}

async function loadTransactions() {
    transactions = await apiCall('/transactions') || [];
    return transactions;
}

async function loadAllocations() {
    const period = getCurrentPeriod();
    allocations = await apiCall(`/allocations?period=${period}`) || [];
    return allocations;
}

async function loadReadyToAssign() {
    const period = getCurrentPeriod();
    const data = await apiCall(`/allocations/ready-to-assign?period=${period}`);
    return data?.ready_to_assign || 0;
}

async function loadAllocationSummary() {
    const period = getCurrentPeriod();
    return await apiCall(`/allocations/summary?period=${period}`) || [];
}

async function loadAccountSummary() {
    return await apiCall('/accounts/summary');
}

// View management - REMOVED (Budget is now the only main view)

// Budget view
async function loadBudgetView() {
    document.getElementById('current-month').textContent = formatMonthYear();

    try {
        await Promise.all([loadCategories(), loadCategoryGroups(), loadAllocations()]);
        const summaryData = await loadAllocationSummary();

        // Extract ready_to_assign and categories from the response
        const readyToAssign = summaryData?.ready_to_assign || 0;
        const summary = summaryData?.categories || [];

        // Update Ready to Assign display with appropriate color
        const readyToAssignEl = document.getElementById('ready-to-assign');
        const readyToAssignBox = document.getElementById('ready-to-assign-box');
        const readyToAssignCheckmark = document.getElementById('ready-to-assign-checkmark');
        const readyToAssignMessage = document.getElementById('ready-to-assign-message');

        readyToAssignEl.textContent = formatCurrency(readyToAssign);

        if (readyToAssign === 0) {
            // All money assigned - show green with checkmark
            readyToAssignEl.className = 'text-3xl font-bold text-green-600 dark:text-green-400';
            readyToAssignBox.className = 'bg-green-50 dark:bg-green-900/30 border border-green-200 dark:border-green-800 rounded-lg p-4 mb-6 transition-colors';
            readyToAssignCheckmark.className = 'text-3xl text-green-600 dark:text-green-400';
            readyToAssignMessage.textContent = 'All money assigned - good to go!';
        } else if (readyToAssign < 0) {
            // Overspent - show red
            readyToAssignEl.className = 'text-3xl font-bold text-red-600 dark:text-red-400';
            readyToAssignBox.className = 'bg-red-50 dark:bg-red-900/30 border border-red-200 dark:border-red-800 rounded-lg p-4 mb-6 transition-colors';
            readyToAssignCheckmark.className = 'text-3xl hidden';
            readyToAssignMessage.textContent = 'Over-allocated! Adjust your budget.';
        } else {
            // Has money to assign - show blue
            readyToAssignEl.className = 'text-3xl font-bold text-blue-600 dark:text-blue-400';
            readyToAssignBox.className = 'bg-blue-50 dark:bg-blue-900/30 border border-blue-200 dark:border-blue-800 rounded-lg p-4 mb-6 transition-colors';
            readyToAssignCheckmark.className = 'text-3xl hidden';
            readyToAssignMessage.textContent = 'Money available to allocate to categories';
        }

        const budgetCategories = document.getElementById('budget-categories');

        if (categories.length === 0) {
            budgetCategories.innerHTML = `
                <div class="text-center py-12">
                    <p class="text-gray-500 dark:text-gray-400 mb-4">No expense categories yet.</p>
                    <button onclick="showAddCategoryModal()" class="btn-primary">Create Your First Category</button>
                </div>
            `;
            return;
        }

        // Render category groups
        budgetCategories.innerHTML = renderBudgetWithGroups(summary);

        // Initialize drag-and-drop after rendering
        initializeBudgetDragDrop();

        // Update expand/collapse all button text
        updateExpandCollapseButton();
    } catch (error) {
        console.error('Failed to load budget view:', error);
    }
}

function renderBudgetWithGroups(summary) {
    let html = '';

    // Sort groups by display order
    const sortedGroups = [...categoryGroups].sort((a, b) => a.display_order - b.display_order);

    // Initialize collapsed state: collapse all groups by default on first load
    // Only initialize if we haven't set any state yet (use a flag to track first load)
    if (typeof window.budgetGroupsInitialized === 'undefined') {
        window.budgetGroupsInitialized = true;
        collapsedGroups.clear();
        sortedGroups.forEach(group => collapsedGroups.add(group.id));
    }

    // Render each group (including empty ones)
    for (const group of sortedGroups) {
        const groupCategories = categories.filter(c => c.group_id === group.id);
        html += renderGroupSection(group, groupCategories, summary);
    }

    return html;
}

function renderGroupSection(group, groupCategories, summary) {
    // Collapsible groups feature
    const isCollapsed = collapsedGroups.has(group.id);
    const chevron = isCollapsed ? '▶' : '▼';
    const contentDisplay = isCollapsed ? 'style="display: none;"' : '';

    // Check if this is the Credit Card Payments group (protected from user modifications)
    const isCreditCardPaymentsGroup = group.name === 'Credit Card Payments';

    const categoriesHtml = groupCategories.length > 0
        ? groupCategories.map(cat => renderBudgetCategory(cat, summary)).join('')
        : (isCreditCardPaymentsGroup
            ? '<div class="text-gray-400 dark:text-gray-500 text-sm p-4 border-2 border-dashed border-gray-200 dark:border-gray-700 rounded text-center">Payment categories are automatically created for credit card accounts</div>'
            : '<div class="text-gray-400 dark:text-gray-500 text-sm p-4 border-2 border-dashed border-gray-200 dark:border-gray-700 rounded text-center">Drag categories here</div>');

    // Conditionally render edit functionality and delete button
    const groupNameHtml = isCreditCardPaymentsGroup
        ? `<span class="px-2 py-1 -mx-2 -my-1 inline-block">${group.name}</span>`
        : `<span class="cursor-pointer hover:bg-gray-200 dark:hover:bg-gray-600 rounded px-2 py-1 -mx-2 -my-1 inline-block no-drag"
                 onclick="event.stopPropagation(); startGroupNameEdit('${group.id}', '${group.name.replace(/'/g, "\\'")}')"
                 title="Click to edit group name">${group.name}</span>`;

    const deleteButtonHtml = isCreditCardPaymentsGroup
        ? '<div class="ml-3 w-5 h-5"></div>' // Empty spacer to maintain layout
        : `<button onclick="event.stopPropagation(); deleteGroup('${group.id}');"
                   class="ml-3 w-5 h-5 flex items-center justify-center text-red-600 dark:text-red-400 hover:text-red-800 dark:hover:text-red-300 hover:bg-red-50 dark:hover:bg-red-900/30 rounded no-drag transition-colors"
                   style="font-size: 12px;"
                   title="Delete group">✕</button>`;

    // Only show "+ Add Category" button for non-Credit Card Payments groups
    const addCategoryButton = isCreditCardPaymentsGroup
        ? ''
        : `<button onclick="showAddCategoryInline('${group.id}', event);" class="mt-2 w-full text-sm text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 hover:bg-blue-50 dark:hover:bg-white/5 rounded px-3 py-2 border border-dashed border-blue-300 dark:border-blue-600 transition">+ Add Category</button>`;

    return `
        <div class="budget-group mb-2" data-group-id="${group.id}" ${isCreditCardPaymentsGroup ? 'data-auto-managed="true"' : ''}>
            <div class="flex justify-between items-center mb-1 mx-px p-2 bg-gray-100 dark:bg-gray-700 rounded transition cursor-pointer hover:bg-gray-200 dark:hover:bg-gray-600"
                 onclick="toggleGroupCollapse('${group.id}')">
                <div class="flex items-center gap-2 flex-1">
                    <span class="collapse-icon text-gray-600 dark:text-gray-400 select-none" style="font-size: 10px; width: 12px;">${chevron}</span>
                    <span class="drag-handle text-gray-400 dark:text-gray-500 cursor-move hover:text-gray-600 dark:hover:text-gray-300 transition no-drag text-xs" title="Drag to reorder" onclick="event.stopPropagation()">⋮⋮</span>
                    <h3 class="text-base font-semibold text-gray-700 dark:text-gray-300 flex-1">
                        ${groupNameHtml}
                    </h3>
                </div>
                <div class="flex gap-4 items-center">
                    <!-- Invisible spacers to align with category columns -->
                    <div class="w-20"></div>
                    <div class="w-20"></div>
                    <div class="w-20"></div>
                    ${deleteButtonHtml}
                </div>
            </div>
            <div class="group-content" ${contentDisplay}>
                <div class="group-categories space-y-1 min-h-[40px]" data-group-id="${group.id}">
                    ${categoriesHtml}
                </div>
                ${addCategoryButton}
            </div>
        </div>
    `;
}

function renderBudgetCategory(category, summary) {
    const allocation = allocations.find(a => a.category_id === category.id);
    const summaryItem = summary.find(s => s.category?.id === category.id);

    const allocated = allocation?.amount || 0;
    const spent = summaryItem?.activity ? -summaryItem.activity : 0;
    const available = summaryItem?.available || (allocated - spent);
    const availableClass = available >= 0 ? 'text-green-600' : 'text-red-600';

    const isPaymentCategory = category.payment_for_account_id != null;
    const isUnderfunded = summaryItem?.underfunded && summaryItem.underfunded > 0;

    // For payment categories, get the credit card account balance
    let cardBalanceDisplay = '';
    if (isPaymentCategory) {
        const creditCardAccount = accounts.find(acc => acc.id === category.payment_for_account_id);
        if (creditCardAccount) {
            const cardBalance = Math.abs(creditCardAccount.balance);
            const balanceClass = creditCardAccount.balance < 0 ? 'text-red-600 dark:text-red-400' : 'text-green-600 dark:text-green-400';
            cardBalanceDisplay = `
                <div class="text-right">
                    <div class="text-xs text-gray-500 dark:text-gray-400">Card Balance</div>
                    <div class="font-medium text-sm ${balanceClass}">${formatCurrency(cardBalance)}</div>
                </div>`;
        }
    }

    const allocatedDisplay = isPaymentCategory
        ? `<div class="font-medium text-sm text-gray-800 dark:text-gray-100" title="Auto-allocated">${formatCurrency(allocated)}</div>`
        : `<div class="font-medium text-sm text-gray-800 dark:text-gray-100 cursor-pointer hover:bg-blue-50 dark:hover:bg-blue-900/30 rounded px-2 py-1 -mx-2 -my-1 no-drag"
                onclick="event.stopPropagation(); startInlineEdit('${category.id}', '${category.name.replace(/'/g, "\\'")}', ${allocated})"
                title="Click to edit">${formatCurrency(allocated)}</div>`;

    const underfundedCategories = summaryItem?.underfunded_categories || [];
    const categoriesText = underfundedCategories.length > 0
        ? `<div class="text-red-600 dark:text-red-400 text-xs mt-0.5">Categories: ${underfundedCategories.join(', ')}</div>`
        : '';

    const underfundedWarning = isUnderfunded
        ? `<div class="mt-1 p-1.5 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded text-xs">
            <div class="flex justify-between items-center">
                <div>
                    <span class="text-red-600 dark:text-red-400 font-medium">⚠️ Underfunded - Need ${formatCurrency(summaryItem.underfunded)} more</span>
                    ${categoriesText}
                </div>
                <button onclick="event.stopPropagation(); coverUnderfunded('${category.id}', '${category.name.replace(/'/g, "\\'")}', ${summaryItem.underfunded});"
                        class="ml-2 px-3 py-1 bg-blue-600 hover:bg-blue-700 text-white rounded text-xs font-medium no-drag transition-colors"
                        title="Allocate funds to cover the shortfall">
                    Allocate to Cover
                </button>
            </div>
        </div>` : '';

    const deleteButton = isPaymentCategory
        ? '<div class="ml-3 w-5 h-5"></div>' // Spacer to maintain alignment
        : `<button onclick="event.stopPropagation(); deleteCategory('${category.id}', '${category.name.replace(/'/g, "\\'")}');"
                   class="ml-3 w-5 h-5 flex items-center justify-center text-red-600 dark:text-red-400 hover:text-red-800 dark:hover:text-red-300 hover:bg-red-50 dark:hover:bg-red-900/30 rounded no-drag transition-colors"
                   style="font-size: 12px;"
                   title="Delete category">✕</button>`;

    return `
        <div class="budget-category border border-gray-200 dark:border-gray-700 rounded p-2 bg-white dark:bg-gray-800 cursor-move ${isPaymentCategory ? 'bg-orange-50 dark:bg-orange-900/20' : ''}"
             data-category-id="${category.id}"
             data-payment-category="${isPaymentCategory}"
             data-group-id="${category.group_id || 'ungrouped'}">
            <div class="flex justify-between items-center">
                <div class="flex items-center gap-2 flex-1">
                    <span class="text-gray-400 dark:text-gray-500 text-xs">⋮⋮</span>
                    <div class="w-2 h-2 rounded-full flex-shrink-0" style="background-color: ${category.color || '#3b82f6'}"></div>
                    <div class="flex-1">
                        <div class="font-medium text-sm text-gray-800 dark:text-gray-100">${category.name}</div>
                    </div>
                </div>
                <div class="flex gap-4 items-center">
                    ${cardBalanceDisplay}
                    <div class="text-right">
                        <div class="text-xs text-gray-500 dark:text-gray-400">Allocated</div>
                        ${allocatedDisplay}
                    </div>
                    <div class="text-right">
                        <div class="text-xs text-gray-500 dark:text-gray-400">${isPaymentCategory ? 'Paid' : 'Spent'}</div>
                        <div class="font-medium text-sm text-gray-800 dark:text-gray-100">${formatCurrency(spent)}</div>
                    </div>
                    <div class="text-right min-w-[80px]">
                        <div class="text-xs text-gray-500 dark:text-gray-400">Available</div>
                        <div class="font-semibold text-sm ${availableClass}">${formatCurrency(available)}</div>
                    </div>
                    ${deleteButton}
                </div>
            </div>
            ${underfundedWarning}
        </div>
    `;
}

function initializeBudgetDragDrop() {
    // Make groups sortable
    const budgetContainer = document.getElementById('budget-categories');
    if (budgetContainer && window.Sortable) {
        new Sortable(budgetContainer, {
            animation: 150,
            handle: '.drag-handle',
            ghostClass: 'opacity-50',
            onEnd: async function(evt) {
                await updateGroupOrder();
            }
        });

        // Make categories within each group sortable
        document.querySelectorAll('.group-categories').forEach(groupEl => {
            new Sortable(groupEl, {
                group: 'categories',
                animation: 150,
                ghostClass: 'opacity-50',
                filter: '.no-drag',
                preventOnFilter: false,
                onMove: function(evt) {
                    // Prevent moving credit card payment categories out of their group
                    const isPaymentCategory = evt.dragged.dataset.paymentCategory === 'true';
                    const currentGroupId = evt.from.dataset.groupId;
                    const targetGroupId = evt.to.dataset.groupId;

                    if (isPaymentCategory && currentGroupId !== targetGroupId) {
                        // Show a brief toast message
                        showToast('Credit card payment categories cannot be moved to other groups', 'error');
                        return false; // Prevent the move
                    }

                    // Prevent moving ANY categories into the Credit Card Payments group
                    const targetGroup = evt.to.closest('.budget-group');
                    const isTargetAutoManaged = targetGroup && targetGroup.dataset.autoManaged === 'true';

                    if (isTargetAutoManaged && currentGroupId !== targetGroupId) {
                        showToast('Cannot add categories to the Credit Card Payments group - it is auto-managed', 'error');
                        return false; // Prevent the move
                    }

                    return true; // Allow the move
                },
                onEnd: async function(evt) {
                    const categoryId = evt.item.dataset.categoryId;
                    const newGroupId = evt.to.dataset.groupId;
                    await updateCategoryGroup(categoryId, newGroupId);
                }
            });
        });
    }
}

async function updateGroupOrder() {
    const groups = [...document.querySelectorAll('.budget-group[data-group-id]')];
    for (let i = 0; i < groups.length; i++) {
        const groupId = groups[i].dataset.groupId;
        try {
            await apiCall(`/category-groups/${groupId}`, {
                method: 'PUT',
                body: JSON.stringify({ display_order: i })
            });
        } catch (error) {
            console.error('Failed to update group order:', error);
        }
    }
}

async function updateCategoryGroup(categoryId, groupId) {
    try {
        if (!groupId) {
            showToast('Categories must belong to a group', 'error');
            loadBudgetView(); // Reload to reset UI
            return;
        }
        await apiCall('/category-groups/assign', {
            method: 'POST',
            body: JSON.stringify({ category_id: categoryId, group_id: groupId })
        });
        showToast('Category moved successfully!');
    } catch (error) {
        console.error('Failed to update category group:', error);
        loadBudgetView(); // Reload on error
    }
}

async function showAddGroupInline() {
    const name = prompt('Enter group name (e.g., Housing, Transportation):');
    if (!name) return;

    try {
        const maxOrder = Math.max(0, ...categoryGroups.map(g => g.display_order));
        await apiCall('/category-groups', {
            method: 'POST',
            body: JSON.stringify({
                name,
                description: '',
                display_order: maxOrder + 1
            })
        });
        await loadCategoryGroups();
        loadBudgetView();
        showToast('Group created successfully!');
    } catch (error) {
        console.error('Failed to create group:', error);
    }
}

async function deleteGroup(groupId) {
    if (!confirm('Delete this group? You must move or delete all categories in this group first.')) return;

    try {
        await apiCall(`/category-groups/${groupId}`, { method: 'DELETE' });
        await loadCategoryGroups();
        loadBudgetView();
        showToast('Group deleted successfully!');
    } catch (error) {
        console.error('Failed to delete group:', error);
        showToast('Cannot delete group: it still contains categories. Please move or delete all categories first.', 'error');
    }
}

function toggleGroupCollapse(groupId) {
    if (collapsedGroups.has(groupId)) {
        collapsedGroups.delete(groupId);
    } else {
        collapsedGroups.add(groupId);
    }
    updateExpandCollapseButton();
    loadBudgetView();
}

function toggleExpandCollapseAll() {
    // If all groups are collapsed, expand all; otherwise collapse all
    const allCollapsed = categoryGroups.length > 0 && collapsedGroups.size === categoryGroups.length;

    if (allCollapsed) {
        // Expand all
        collapsedGroups.clear();
    } else {
        // Collapse all
        collapsedGroups.clear();
        categoryGroups.forEach(group => collapsedGroups.add(group.id));
    }

    updateExpandCollapseButton();
    loadBudgetView();
}

function updateExpandCollapseButton() {
    const button = document.getElementById('expand-collapse-btn');
    if (!button) return;

    const allCollapsed = categoryGroups.length > 0 && collapsedGroups.size === categoryGroups.length;
    button.textContent = allCollapsed ? 'Expand All' : 'Collapse All';
}

// Make functions available globally for onclick handlers
window.toggleGroupCollapse = toggleGroupCollapse;
window.toggleExpandCollapseAll = toggleExpandCollapseAll;

// Inline category management functions
function showAddCategoryInline(groupId, event) {
    // Validate groupId is required
    if (!groupId || groupId === 'null') {
        showToast('Cannot add category: group is required', 'error');
        return;
    }

    const colors = [
        { hex: '#f97316', name: 'Orange' },
        { hex: '#3b82f6', name: 'Blue' },
        { hex: '#10b981', name: 'Green' },
        { hex: '#a855f7', name: 'Purple' },
        { hex: '#ef4444', name: 'Red' },
        { hex: '#ec4899', name: 'Pink' },
        { hex: '#eab308', name: 'Yellow' },
        { hex: '#6366f1', name: 'Indigo' },
        { hex: '#14b8a6', name: 'Teal' },
        { hex: '#6b7280', name: 'Gray' }
    ];

    const colorButtons = colors.map(color =>
        `<button type="button" onclick="selectInlineColor('${color.hex}')"
                 class="inline-color-btn w-6 h-6 rounded-full hover:ring-2 hover:ring-blue-400 transition"
                 style="background-color: ${color.hex}"
                 data-color="${color.hex}"
                 title="${color.name}"></button>`
    ).join('');

    const groupSelector = `<input type="hidden" id="inline-category-group" value="${groupId}">`;

    const formHtml = `
        <div id="inline-category-form" class="bg-blue-50 dark:bg-blue-900/30 border-2 border-blue-300 dark:border-blue-600 rounded-lg p-4 mt-2">
            <h4 class="font-semibold mb-3 text-gray-800 dark:text-gray-100">Add New Category</h4>
            <div class="space-y-3">
                <div>
                    <label class="block text-sm font-medium mb-1 text-gray-700 dark:text-gray-300">Name *</label>
                    <input type="text" id="inline-category-name" class="w-full border border-gray-300 dark:border-gray-600 rounded px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:outline-none bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100" placeholder="Category name" required>
                </div>
                <div>
                    <label class="block text-sm font-medium mb-1 text-gray-700 dark:text-gray-300">Color</label>
                    <div class="flex gap-2 flex-wrap">
                        ${colorButtons}
                    </div>
                    <input type="hidden" id="inline-category-color" value="#3b82f6">
                </div>
                ${groupSelector}
                <div class="flex gap-2">
                    <button type="button" onclick="saveInlineCategory()" class="btn-primary text-sm">Add Category</button>
                    <button type="button" onclick="cancelInlineCategory()" class="btn-secondary text-sm">Cancel</button>
                </div>
            </div>
        </div>
    `;

    // Remove any existing form
    const existing = document.getElementById('inline-category-form');
    if (existing) existing.remove();

    // Find the right place to insert the form
    const targetButton = event ? event.target : document.querySelector(`button[onclick*="showAddCategoryInline"]`);
    if (targetButton) {
        targetButton.insertAdjacentHTML('beforebegin', formHtml);
    } else {
        // Fallback: append to the appropriate group container
        const targetContainer = document.querySelector(`.group-categories[data-group-id="${groupId}"]`);
        if (targetContainer) {
            targetContainer.insertAdjacentHTML('afterend', formHtml);
        }
    }

    // Focus on name input
    setTimeout(() => {
        const nameInput = document.getElementById('inline-category-name');
        if (nameInput) nameInput.focus();
    }, 50);

    // Highlight default color
    selectInlineColor('#3b82f6');
}

function selectInlineColor(color) {
    // Remove selection from all buttons
    document.querySelectorAll('.inline-color-btn').forEach(btn => {
        btn.classList.remove('ring-2', 'ring-blue-600', 'dark:ring-blue-400');
    });

    // Add selection to clicked button
    const selectedBtn = document.querySelector(`.inline-color-btn[data-color="${color}"]`);
    if (selectedBtn) {
        selectedBtn.classList.add('ring-2', 'ring-blue-600', 'dark:ring-blue-400');
    }

    // Update hidden input
    const colorInput = document.getElementById('inline-category-color');
    if (colorInput) {
        colorInput.value = color;
    }
}

async function saveInlineCategory() {
    const name = document.getElementById('inline-category-name').value.trim();
    const color = document.getElementById('inline-category-color').value;
    const groupId = document.getElementById('inline-category-group').value;

    if (!name) {
        showToast('Please enter a category name', 'error');
        return;
    }

    // Validate groupId is required
    if (!groupId || groupId === '' || groupId === 'null') {
        showToast('Group is required for all categories', 'error');
        return;
    }

    try {
        const categoryData = {
            name,
            group_id: groupId,
            color,
            description: ''
        };

        await apiCall('/categories', {
            method: 'POST',
            body: JSON.stringify(categoryData)
        });

        showToast('Category added!');
        cancelInlineCategory();
        await loadCategories();
        await loadBudgetView();
    } catch (error) {
        console.error('Failed to add category:', error);
        showToast('Failed to add category', 'error');
    }
}

function cancelInlineCategory() {
    const form = document.getElementById('inline-category-form');
    if (form) form.remove();
}

async function deleteCategory(categoryId, categoryName) {
    if (!confirm(`Delete category "${categoryName}"?\n\nThis will remove the category and unassign it from all transactions.`)) {
        return;
    }

    try {
        await apiCall(`/categories/${categoryId}`, { method: 'DELETE' });
        showToast('Category deleted!');
        await loadCategories();
        await loadBudgetView();
        await loadSidebar(); // Refresh sidebar to update recent transactions
    } catch (error) {
        console.error('Failed to delete category:', error);
        showToast('Failed to delete category', 'error');
    }
}

// Make functions globally available
window.showAddCategoryInline = showAddCategoryInline;
window.selectInlineColor = selectInlineColor;
window.saveInlineCategory = saveInlineCategory;
window.cancelInlineCategory = cancelInlineCategory;
window.deleteCategory = deleteCategory;

// Accounts view
async function loadAccountsView() {
    try {
        await loadAccounts();
        const summary = await loadAccountSummary();

        if (summary) {
            document.getElementById('total-balance').textContent = formatCurrency(summary.total_balance);
        }

        const accountsList = document.getElementById('accounts-list');

        if (accounts.length === 0) {
            accountsList.innerHTML = `
                <div class="text-center py-12">
                    <p class="text-gray-500 dark:text-gray-400 mb-4">No accounts yet. Create one to start tracking your money!</p>
                    <button onclick="showAddAccountModal()" class="btn-primary">Create Your First Account</button>
                </div>
            `;
            return;
        }

        accountsList.innerHTML = accounts.map(account => {
            const isCreditCard = account.type === 'credit';
            const balanceClass = account.balance >= 0 ? 'text-green-600' : 'text-red-600';

            // For credit cards, show "Owe $X.XX" instead of negative amount
            let balanceDisplay;
            if (isCreditCard && account.balance < 0) {
                balanceDisplay = `<span class="text-sm text-gray-500 dark:text-gray-400">Owe </span>${formatCurrency(Math.abs(account.balance))}`;
            } else {
                balanceDisplay = formatCurrency(account.balance);
            }

            return `
                <div class="border border-gray-200 dark:border-gray-700 rounded-lg p-4 hover:shadow-md transition-shadow">
                    <div class="flex justify-between items-center">
                        <div>
                            <div class="font-semibold text-gray-800 dark:text-gray-100">${account.name}</div>
                            <div class="text-sm text-gray-500 dark:text-gray-400 capitalize">${account.type}</div>
                        </div>
                        <div class="text-right">
                            <div class="text-xl font-bold ${balanceClass}">${balanceDisplay}</div>
                        </div>
                    </div>
                </div>
            `;
        }).join('');
    } catch (error) {
        console.error('Failed to load accounts view:', error);
    }
}

// Transactions view
async function loadTransactionsView() {
    try {
        await loadTransactions();
        await loadAccounts();
        await loadCategories();

        const transactionsList = document.getElementById('transactions-list');

        if (transactions.length === 0) {
            transactionsList.innerHTML = `
                <div class="text-center py-12">
                    <p class="text-gray-500 dark:text-gray-400 mb-4">No transactions yet.</p>
                    <button onclick="showAddTransactionModal()" class="btn-primary">Add Your First Transaction</button>
                </div>
            `;
            return;
        }

        // Sort by date descending
        const sortedTransactions = [...transactions].sort((a, b) =>
            new Date(b.date) - new Date(a.date)
        );

        transactionsList.innerHTML = sortedTransactions.map(transaction => {
            const account = accounts.find(a => a.id === transaction.account_id);
            const category = categories.find(c => c.id === transaction.category_id);
            const amountClass = transaction.amount >= 0 ? 'text-green-600' : 'text-red-600';
            const sign = transaction.amount >= 0 ? '+' : '';

            // Handle transfer transactions
            let transactionInfo = '';
            if (transaction.type === 'transfer') {
                const toAccount = accounts.find(a => a.id === transaction.transfer_to_account_id);
                transactionInfo = `${formatDate(transaction.date)} • Transfer: ${account?.name || 'Unknown'} → ${toAccount?.name || 'Unknown'}`;
            } else {
                transactionInfo = `${formatDate(transaction.date)} • ${account?.name || 'Unknown'}${category ? ' • ' + category.name : ''}`;
            }

            return `
                <div class="border border-gray-200 dark:border-gray-700 rounded-lg p-4 hover:shadow-md transition-shadow">
                    <div class="flex justify-between items-center">
                        <div class="flex-1">
                            <div class="flex items-center gap-2">
                                ${category ? `<div class="w-2 h-2 rounded-full" style="background-color: ${category.color || '#gray'}"></div>` : ''}
                                <div class="font-semibold text-gray-800 dark:text-gray-100">${transaction.description || 'Transaction'}</div>
                            </div>
                            <div class="text-sm text-gray-500 dark:text-gray-400 mt-1">
                                ${transactionInfo}
                            </div>
                        </div>
                        <div class="text-right">
                            <div class="text-xl font-bold ${amountClass}">${sign}${formatCurrency(Math.abs(transaction.amount))}</div>
                        </div>
                    </div>
                </div>
            `;
        }).join('');
    } catch (error) {
        console.error('Failed to load transactions view:', error);
    }
}

// Categories view
async function loadCategoriesView() {
    try {
        await loadCategories();

        // Filter out payment categories (auto-created for credit cards)
        const userCategories = categories.filter(c => !c.payment_for_account_id);
        const categoriesList = document.getElementById('categories-list');

        if (userCategories.length === 0) {
            categoriesList.innerHTML = '<div class="text-gray-500 dark:text-gray-400 text-center py-4">No categories yet.</div>';
        } else {
            // Show flat list of categories (groups are managed on budget page)
            categoriesList.innerHTML = userCategories.map(category => renderCategoryCard(category)).join('');
        }
    } catch (error) {
        console.error('Failed to load categories view:', error);
    }
}

function renderCategoriesByGroups(categoriesList, groups) {
    let html = '';

    // Render groups with their categories
    for (const group of groups) {
        const groupCategories = categoriesList.filter(c => c.group_id === group.id);
        if (groupCategories.length > 0) {
            html += `
                <div class="mb-6">
                    <h3 class="text-lg font-semibold text-gray-700 mb-3">${group.name}</h3>
                    ${group.description ? `<p class="text-sm text-gray-500 mb-3">${group.description}</p>` : ''}
                    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                        ${groupCategories.map(category => renderCategoryCard(category)).join('')}
                    </div>
                </div>
            `;
        }
    }

    // Render ungrouped categories
    const ungroupedCategories = categoriesList.filter(c => !c.group_id);
    if (ungroupedCategories.length > 0) {
        html += `
            <div class="mb-6">
                <h3 class="text-lg font-semibold text-gray-700 mb-3">Ungrouped</h3>
                <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                    ${ungroupedCategories.map(category => renderCategoryCard(category)).join('')}
                </div>
            </div>
        `;
    }

    return html || '<div class="text-gray-500 text-center py-4">No categories yet.</div>';
}

function renderCategoryCard(category) {
    return `
        <div class="border border-gray-200 dark:border-gray-700 rounded-lg p-4 hover:shadow-md transition-shadow">
            <div class="flex items-center gap-3">
                <div class="w-4 h-4 rounded-full flex-shrink-0" style="background-color: ${category.color || '#3b82f6'}"></div>
                <div class="flex-1">
                    <div class="font-semibold text-gray-800 dark:text-gray-100">${category.name}</div>
                    ${category.description ? `<div class="text-sm text-gray-500 dark:text-gray-400">${category.description}</div>` : ''}
                </div>
            </div>
        </div>
    `;
}

// Month navigation
function changeMonth(delta) {
    currentMonth.setMonth(currentMonth.getMonth() + delta);
    loadBudgetView();
}

// Modal functions
function showModal(modalId) {
    document.getElementById(modalId).classList.add('active');
}

function closeModal(modalId) {
    document.getElementById(modalId).classList.remove('active');
}

async function showAddTransactionModal() {
    await loadAccounts();
    await loadCategories();

    if (accounts.length === 0) {
        showToast('Please create an account first', 'error');
        showAddAccountModal();
        return;
    }

    if (categories.length === 0) {
        showToast('Please create a category first', 'error');
        showAddCategoryModal();
        return;
    }

    // Populate account and category dropdowns
    const accountSelect = document.getElementById('transaction-account');
    const categorySelect = document.getElementById('transaction-category');

    accountSelect.innerHTML = '<option value="">Select account...</option>' +
        accounts.map(a => `<option value="${a.id}">${a.name}</option>`).join('');

    // Filter out payment categories (auto-created for credit cards)
    const userCategories = categories.filter(c => !c.payment_for_account_id);
    categorySelect.innerHTML = '<option value="">Select category...</option>' +
        userCategories.map(c => `<option value="${c.id}">${c.name}</option>`).join('');

    // Set default date to today
    const today = new Date().toISOString().split('T')[0];
    document.getElementById('transaction-date').value = today;

    // Reset amount label and hint to defaults
    const amountLabel = document.getElementById('amount-label');
    const amountHint = document.getElementById('amount-hint');
    const amountInput = document.getElementById('transaction-amount');
    amountLabel.textContent = 'Amount *';
    amountHint.textContent = 'Enter positive for inflow, negative for outflow';
    amountInput.placeholder = '-45.67 for outflow, 2500.00 for inflow';
    amountInput.removeAttribute('min');

    showModal('transaction-modal');
}

async function showAddTransferModal() {
    await loadAccounts();

    if (accounts.length < 2) {
        showToast('You need at least 2 accounts to make a transfer', 'error');
        return;
    }

    // Populate account dropdowns
    const fromAccountSelect = document.getElementById('transfer-from-account');
    const toAccountSelect = document.getElementById('transfer-to-account');

    const accountOptions = accounts.map(a => `<option value="${a.id}">${a.name}</option>`).join('');
    fromAccountSelect.innerHTML = '<option value="">Select account...</option>' + accountOptions;
    toAccountSelect.innerHTML = '<option value="">Select account...</option>' + accountOptions;

    // Set default date to today
    const today = new Date().toISOString().split('T')[0];
    document.getElementById('transfer-date').value = today;

    showModal('transfer-modal');
}

function showAddAccountModal() {
    document.getElementById('account-form').reset();
    showModal('account-modal');
}

function showAddCategoryModal() {
    document.getElementById('category-form').reset();

    // Reset color swatches to default (blue)
    document.querySelectorAll('.color-swatch').forEach(swatch => {
        swatch.classList.remove('selected');
        swatch.querySelector('.color-check').classList.add('hidden');
    });
    const defaultSwatch = document.querySelector('.color-swatch[data-color="#3b82f6"]');
    if (defaultSwatch) {
        defaultSwatch.classList.add('selected');
        defaultSwatch.querySelector('.color-check').classList.remove('hidden');
    }
    document.getElementById('category-color').value = '#3b82f6';

    // Populate group dropdown
    const groupSelect = document.getElementById('category-group');
    groupSelect.innerHTML = '<option value="">Select group...</option>';
    categoryGroups.forEach(group => {
        const option = document.createElement('option');
        option.value = group.id;
        option.textContent = group.name;
        groupSelect.appendChild(option);
    });

    showModal('category-modal');
}

function showAllocateModal(categoryId, categoryName, currentAmount = 0) {
    document.getElementById('allocation-category-id').value = categoryId;
    document.getElementById('allocation-current-amount').value = currentAmount; // Store in cents
    document.getElementById('allocation-category-name').textContent = categoryName;
    document.getElementById('allocation-amount').value = (currentAmount / 100).toFixed(2);
    document.getElementById('allocation-notes').value = '';
    showModal('allocation-modal');
}

// Helper function to parse allocation input with math operators
function parseAllocationInput(inputValue, currentAmountInCents) {
    const trimmed = inputValue.trim();

    // Check if input starts with a math operator
    const operatorMatch = trimmed.match(/^([+\-*/])(.+)$/);

    if (operatorMatch) {
        const operator = operatorMatch[1];
        const operand = parseFloat(operatorMatch[2]);

        if (isNaN(operand)) {
            return { valid: false, error: 'Invalid number after operator' };
        }

        const currentAmountInDollars = currentAmountInCents / 100;
        let result;

        switch (operator) {
            case '+':
                result = currentAmountInDollars + operand;
                break;
            case '-':
                result = currentAmountInDollars - operand;
                break;
            case '*':
                result = currentAmountInDollars * operand;
                break;
            case '/':
                if (operand === 0) {
                    return { valid: false, error: 'Cannot divide by zero' };
                }
                result = currentAmountInDollars / operand;
                break;
        }

        if (result < 0) {
            return { valid: false, error: 'Result cannot be negative' };
        }

        return { valid: true, amountInCents: Math.round(result * 100) };
    }

    // No operator, treat as absolute value
    const amount = parseFloat(trimmed);

    if (isNaN(amount) || amount < 0) {
        return { valid: false, error: 'Please enter a valid amount' };
    }

    return { valid: true, amountInCents: Math.round(amount * 100) };
}

// Inline editing for budget allocation
async function startInlineEdit(categoryId, categoryName, currentAmount) {
    // Find the element that was clicked
    const clickedElement = event.target;
    const container = clickedElement.parentElement;

    // Store original content
    const originalContent = clickedElement.innerHTML;

    // Create input element
    const input = document.createElement('input');
    input.type = 'text';
    input.value = (currentAmount / 100).toFixed(2);
    input.className = 'w-24 border border-blue-500 dark:border-blue-400 rounded px-2 py-1 text-center font-semibold bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400';
    input.placeholder = 'e.g. +50, -25, 100';

    // Replace content with input
    clickedElement.innerHTML = '';
    clickedElement.appendChild(input);
    input.focus();
    input.select();

    // Function to save the allocation
    const saveAllocation = async () => {
        const result = parseAllocationInput(input.value, currentAmount);

        if (!result.valid) {
            showToast(result.error, 'error');
            clickedElement.innerHTML = originalContent;
            return;
        }

        const amountInCents = result.amountInCents;

        // Only save if the amount changed
        if (amountInCents !== currentAmount) {
            try {
                const period = getCurrentPeriod();
                await apiCall('/allocations', {
                    method: 'POST',
                    body: JSON.stringify({
                        category_id: categoryId,
                        amount: amountInCents,
                        period,
                        notes: ''
                    })
                });

                showToast('Allocation updated!');
                loadBudgetView();
            } catch (error) {
                console.error('Failed to update allocation:', error);
                clickedElement.innerHTML = originalContent;
            }
        } else {
            clickedElement.innerHTML = originalContent;
        }
    };

    // Function to cancel editing
    const cancelEdit = () => {
        clickedElement.innerHTML = originalContent;
    };

    // Handle Enter key to save
    input.addEventListener('keydown', (e) => {
        if (e.key === 'Enter') {
            e.preventDefault();
            saveAllocation();
        } else if (e.key === 'Escape') {
            e.preventDefault();
            cancelEdit();
        }
    });

    // Handle click outside to save
    input.addEventListener('blur', () => {
        setTimeout(() => saveAllocation(), 100);
    });
}

// Cover underfunded payment category
async function coverUnderfunded(categoryId, categoryName, underfundedAmount) {
    const period = getCurrentPeriod();

    // Confirm with the user
    const confirmed = confirm(
        `Allocate ${formatCurrency(underfundedAmount)} to "${categoryName}" to cover the credit card balance?\n\n` +
        `This will use available funds from Ready to Assign.`
    );

    if (!confirmed) {
        return;
    }

    try {
        await apiCall('/allocations/cover-underfunded', {
            method: 'POST',
            body: JSON.stringify({
                category_id: categoryId,
                period: period
            })
        });

        showToast(`Allocated ${formatCurrency(underfundedAmount)} to ${categoryName}!`);
        loadBudgetView();
    } catch (error) {
        console.error('Failed to cover underfunded category:', error);
        showToast(error.message || 'Failed to allocate funds. Check if you have enough available.', 'error');
    }
}

// Inline category name editing
async function startCategoryNameEdit(categoryId, currentName) {
    const clickedElement = event.target;
    const originalContent = clickedElement.innerHTML;

    const input = document.createElement('input');
    input.type = 'text';
    input.value = currentName;
    input.className = 'border border-blue-500 dark:border-blue-400 rounded px-2 py-1 font-semibold bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400';

    clickedElement.innerHTML = '';
    clickedElement.appendChild(input);
    input.focus();
    input.select();

    const saveName = async () => {
        const newName = input.value.trim();

        if (!newName) {
            showToast('Category name cannot be empty', 'error');
            clickedElement.innerHTML = originalContent;
            return;
        }

        if (newName !== currentName) {
            try {
                await apiCall(`/categories/${categoryId}`, {
                    method: 'PUT',
                    body: JSON.stringify({
                        name: newName,
                        color: '', // Will be preserved by backend
                        description: ''
                    })
                });

                showToast('Category name updated!');
                loadBudgetView();
            } catch (error) {
                console.error('Failed to update category name:', error);
                clickedElement.innerHTML = originalContent;
            }
        } else {
            clickedElement.innerHTML = originalContent;
        }
    };

    const cancelEdit = () => {
        clickedElement.innerHTML = originalContent;
    };

    input.addEventListener('keydown', (e) => {
        if (e.key === 'Enter') {
            e.preventDefault();
            saveName();
        } else if (e.key === 'Escape') {
            e.preventDefault();
            cancelEdit();
        }
    });

    input.addEventListener('blur', () => {
        setTimeout(() => saveName(), 100);
    });
}


// Show color picker for category
function showColorPicker(categoryId, currentColor) {
    const colors = [
        { hex: '#f97316', name: 'Orange' },
        { hex: '#3b82f6', name: 'Blue' },
        { hex: '#10b981', name: 'Green' },
        { hex: '#a855f7', name: 'Purple' },
        { hex: '#ef4444', name: 'Red' },
        { hex: '#ec4899', name: 'Pink' },
        { hex: '#eab308', name: 'Yellow' },
        { hex: '#6366f1', name: 'Indigo' },
        { hex: '#14b8a6', name: 'Teal' },
        { hex: '#6b7280', name: 'Gray' }
    ];

    const colorButtons = colors.map(color =>
        `<button onclick="updateCategoryColor('${categoryId}', '${color.hex}');"
                 class="w-8 h-8 rounded-full hover:ring-2 hover:ring-offset-2 hover:ring-blue-400 transition ${color.hex === currentColor ? 'ring-2 ring-blue-600' : ''}"
                 style="background-color: ${color.hex}"
                 title="${color.name}"></button>`
    ).join('');

    const picker = document.createElement('div');
    picker.id = 'color-picker-popup';
    picker.className = 'fixed bg-white border-2 border-gray-300 rounded-lg shadow-xl p-4 z-50';
    picker.style.left = event.pageX + 'px';
    picker.style.top = event.pageY + 'px';
    picker.innerHTML = `
        <div class="mb-2 text-sm font-semibold text-gray-700">Select Color</div>
        <div class="grid grid-cols-5 gap-2 mb-2">
            ${colorButtons}
        </div>
        <button onclick="closeColorPicker()" class="text-xs text-gray-600 hover:text-gray-800 w-full">Cancel</button>
    `;

    // Remove any existing picker
    const existing = document.getElementById('color-picker-popup');
    if (existing) existing.remove();

    document.body.appendChild(picker);

    // Close on click outside
    setTimeout(() => {
        document.addEventListener('click', function closeOnClickOutside(e) {
            if (!picker.contains(e.target)) {
                closeColorPicker();
                document.removeEventListener('click', closeOnClickOutside);
            }
        });
    }, 100);
}

function closeColorPicker() {
    const picker = document.getElementById('color-picker-popup');
    if (picker) picker.remove();
}

async function updateCategoryColor(categoryId, newColor) {
    closeColorPicker();

    try {
        await apiCall(`/categories/${categoryId}`, {
            method: 'PUT',
            body: JSON.stringify({
                name: '', // Will be preserved by backend
                color: newColor,
                description: ''
            })
        });

        showToast('Color updated!');
        loadBudgetView();
    } catch (error) {
        console.error('Failed to update category color:', error);
    }
}
async function startGroupNameEdit(groupId, currentName) {
    const clickedElement = event.target;
    const originalContent = clickedElement.innerHTML;

    const input = document.createElement('input');
    input.type = 'text';
    input.value = currentName;
    input.className = 'border border-blue-500 dark:border-blue-400 rounded px-2 py-1 text-lg font-semibold bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400';

    clickedElement.innerHTML = '';
    clickedElement.appendChild(input);
    input.focus();
    input.select();

    const saveName = async () => {
        const newName = input.value.trim();

        if (!newName) {
            showToast('Group name cannot be empty', 'error');
            clickedElement.innerHTML = originalContent;
            return;
        }

        if (newName !== currentName) {
            try {
                await apiCall(`/category-groups/${groupId}`, {
                    method: 'PUT',
                    body: JSON.stringify({
                        name: newName,
                        description: '',
                        display_order: 0 // Will be preserved by backend
                    })
                });

                showToast('Group name updated!');
                loadBudgetView();
            } catch (error) {
                console.error('Failed to update group name:', error);
                clickedElement.innerHTML = originalContent;
            }
        } else {
            clickedElement.innerHTML = originalContent;
        }
    };

    const cancelEdit = () => {
        clickedElement.innerHTML = originalContent;
    };

    input.addEventListener('keydown', (e) => {
        if (e.key === 'Enter') {
            e.preventDefault();
            saveName();
        } else if (e.key === 'Escape') {
            e.preventDefault();
            cancelEdit();
        }
    });

    input.addEventListener('blur', () => {
        setTimeout(() => saveName(), 100);
    });
}

// Make function globally available
window.startGroupNameEdit = startGroupNameEdit;

// Load uncategorized transactions
async function loadUncategorizedTransactions() {
    try {
        const transactions = await apiCall('/transactions?uncategorized=true');

        // Filter to only show outflows (negative amounts) - inflows don't need categorization
        const outflows = transactions.filter(txn => txn.amount < 0);

        const listContainer = document.getElementById('uncategorized-list');

        if (outflows.length === 0) {
            listContainer.innerHTML = '<p class="text-gray-500 dark:text-gray-400 text-center py-4">No uncategorized transactions</p>';
            return;
        }

        listContainer.innerHTML = `
            <div class="mb-3 flex gap-2">
                <button onclick="selectAllUncategorized()" class="btn-secondary text-sm">Select All</button>
                <button onclick="showCategorizeModal()" class="btn-primary text-sm">Categorize Selected</button>
            </div>
            ${outflows.map(txn => {
                const account = accounts.find(a => a.id === txn.account_id);
                const amountClass = txn.amount >= 0 ? 'text-green-600' : 'text-red-600';
                return `
                    <div class="flex items-center gap-3 p-3 bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 hover:border-blue-300 dark:hover:border-blue-600 transition">
                        <input type="checkbox" class="uncategorized-checkbox" data-transaction-id="${txn.id}">
                        <div class="flex-1 min-w-0">
                            <div class="flex justify-between items-start gap-2">
                                <div class="flex-1 min-w-0">
                                    <div class="font-medium text-gray-800 dark:text-gray-100 truncate">${txn.description || 'No description'}</div>
                                    <div class="text-xs text-gray-500 dark:text-gray-400">${account ? account.name : 'Unknown'} • ${new Date(txn.date).toLocaleDateString()}</div>
                                </div>
                                <div class="font-semibold ${amountClass} whitespace-nowrap">${formatCurrency(txn.amount)}</div>
                            </div>
                        </div>
                    </div>
                `;
            }).join('')}
        `;
    } catch (error) {
        console.error('Failed to load uncategorized transactions:', error);
    }
}

// Select all uncategorized transactions
function selectAllUncategorized() {
    const checkboxes = document.querySelectorAll('.uncategorized-checkbox');
    const allChecked = Array.from(checkboxes).every(cb => cb.checked);
    checkboxes.forEach(cb => cb.checked = !allChecked);
}

// Show categorize modal
function showCategorizeModal() {
    const checkboxes = document.querySelectorAll('.uncategorized-checkbox:checked');
    const selectedIds = Array.from(checkboxes).map(cb => cb.dataset.transactionId);

    if (selectedIds.length === 0) {
        showToast('Please select transactions to categorize', 'error');
        return;
    }

    window.selectedTransactions = selectedIds;
    document.getElementById('categorize-count').textContent = selectedIds.length;

    // Populate category dropdown
    const categorySelect = document.getElementById('categorize-category');
    categorySelect.innerHTML = '<option value="">Select category...</option>' +
        categories.map(cat => `<option value="${cat.id}">${cat.name} (${cat.type})</option>`).join('');

    showModal('categorize-modal');
}

// Load import view
async function loadImportView() {
    // Populate account dropdown
    const accountSelect = document.getElementById('import-account');
    accountSelect.innerHTML = '<option value="">Choose account to import into...</option>' +
        accounts.map(acc => `<option value="${acc.id}">${acc.name} (${acc.type})</option>`).join('');

    // Load uncategorized transactions
    await loadUncategorizedTransactions();
}

// Form submissions
document.addEventListener('DOMContentLoaded', function() {
    // Initialize theme
    initializeTheme();

    // Add listener for account change to update amount hint text
    document.getElementById('transaction-account').addEventListener('change', function() {
        const amountLabel = document.getElementById('amount-label');
        const amountHint = document.getElementById('amount-hint');
        const amountInput = document.getElementById('transaction-amount');
        const selectedAccountId = this.value;

        if (!selectedAccountId) {
            amountLabel.textContent = 'Amount *';
            amountHint.textContent = 'Enter positive for inflow, negative for outflow';
            amountInput.placeholder = '-45.67 for outflow, 2500.00 for inflow';
            return;
        }

        const selectedAccount = accounts.find(a => a.id === selectedAccountId);

        if (selectedAccount && selectedAccount.type === 'credit') {
            amountLabel.textContent = 'How much did you spend? *';
            amountHint.textContent = 'Enter amount spent (we\'ll deduct it automatically). To make a payment, use the Transfer button.';
            amountInput.placeholder = '45.67';
            amountInput.min = '0.01'; // Only allow positive amounts for spending
        } else {
            amountLabel.textContent = 'Amount *';
            amountHint.textContent = 'Enter positive for inflow, negative for outflow';
            amountInput.placeholder = '-45.67 for outflow, 2500.00 for inflow';
            amountInput.removeAttribute('min'); // Allow negative amounts for regular accounts
        }
    });

    // Add listener for amount change to update category requirement
    document.getElementById('transaction-amount').addEventListener('input', function() {
        const categorySelect = document.getElementById('transaction-category');
        const categoryIndicator = document.getElementById('category-required-indicator');
        const amount = parseFloat(this.value);

        if (amount > 0 || isNaN(amount)) {
            // Inflow or empty: category is optional
            categorySelect.removeAttribute('required');
            categoryIndicator.textContent = '';
        } else {
            // Outflow (negative): category is required
            categorySelect.setAttribute('required', 'required');
            categoryIndicator.textContent = '*';
        }
    });

    // Color swatch selection
    document.querySelectorAll('.color-swatch').forEach(swatch => {
        swatch.addEventListener('click', function(e) {
            e.preventDefault();
            // Remove selected class from all swatches
            document.querySelectorAll('.color-swatch').forEach(s => {
                s.classList.remove('selected');
                s.querySelector('.color-check').classList.add('hidden');
            });
            // Add selected class to clicked swatch
            this.classList.add('selected');
            this.querySelector('.color-check').classList.remove('hidden');
            // Update hidden input
            document.getElementById('category-color').value = this.dataset.color;
        });
    });

    // Transaction form
    document.getElementById('transaction-form').addEventListener('submit', async (e) => {
        e.preventDefault();

        const accountId = document.getElementById('transaction-account').value;
        const categoryId = document.getElementById('transaction-category').value;
        let amount = parseFloat(document.getElementById('transaction-amount').value);
        const date = document.getElementById('transaction-date').value;
        const description = document.getElementById('transaction-description').value;

        if (!accountId) {
            showToast('Please select an account', 'error');
            return;
        }

        // For credit card accounts, auto-negate positive amounts (spending)
        // Payments to credit cards should be done via transfers, not this form
        const selectedAccount = accounts.find(a => a.id === accountId);
        if (selectedAccount && selectedAccount.type === 'credit') {
            if (amount < 0) {
                showToast('To make a credit card payment, use the Transfer button instead', 'error');
                return;
            }
            // Auto-negate for spending
            amount = -amount;
        }

        // Category is required for outflows (negative amounts) but optional for inflows
        if (amount < 0 && !categoryId) {
            showToast('Please select a category for outflows', 'error');
            return;
        }

        // Convert amount to cents
        const amountInCents = Math.round(amount * 100);

        try {
            await apiCall('/transactions', {
                method: 'POST',
                body: JSON.stringify({
                    account_id: accountId,
                    category_id: categoryId || null,
                    amount: amountInCents,
                    description: description || 'Transaction',
                    date: new Date(date).toISOString()
                })
            });

            closeModal('transaction-modal');
            document.getElementById('transaction-form').reset();
            showToast('Transaction added successfully!');

            // Reload budget and sidebar
            await loadAccounts();
            await loadBudgetView();
            await loadSidebar();
        } catch (error) {
            console.error('Failed to create transaction:', error);
        }
    });

    // Transfer form
    document.getElementById('transfer-form').addEventListener('submit', async (e) => {
        e.preventDefault();

        const fromAccountId = document.getElementById('transfer-from-account').value;
        const toAccountId = document.getElementById('transfer-to-account').value;
        const amount = parseFloat(document.getElementById('transfer-amount').value);
        const date = document.getElementById('transfer-date').value;
        const description = document.getElementById('transfer-description').value;

        if (!fromAccountId || !toAccountId) {
            showToast('Please select both accounts', 'error');
            return;
        }

        if (fromAccountId === toAccountId) {
            showToast('Cannot transfer to the same account', 'error');
            return;
        }

        // Convert amount to cents
        const amountInCents = Math.round(amount * 100);

        try {
            await apiCall('/transactions/transfer', {
                method: 'POST',
                body: JSON.stringify({
                    from_account_id: fromAccountId,
                    to_account_id: toAccountId,
                    amount: amountInCents,
                    description: description || 'Transfer',
                    date: new Date(date).toISOString()
                })
            });

            closeModal('transfer-modal');
            document.getElementById('transfer-form').reset();
            showToast('Transfer created successfully!');

            // Reload budget and sidebar (including payment category updates)
            await loadAccounts();
            await loadBudgetView();
            await loadSidebar();
        } catch (error) {
            console.error('Failed to create transfer:', error);
        }
    });

    // Account form
    document.getElementById('account-form').addEventListener('submit', async (e) => {
        e.preventDefault();

        const name = document.getElementById('account-name').value;
        const type = document.getElementById('account-type').value;
        const balance = parseFloat(document.getElementById('account-balance').value);

        try {
            await apiCall('/accounts', {
                method: 'POST',
                body: JSON.stringify({
                    name,
                    type,
                    balance: Math.round(balance * 100)
                })
            });

            closeModal('account-modal');
            document.getElementById('account-form').reset();
            showToast('Account created successfully!');

            // Reload accounts, sidebar, and budget view to update RTA
            await loadAccounts();
            await loadSidebar();
            await loadBudgetView();
        } catch (error) {
            console.error('Failed to create account:', error);
        }
    });

    // Category form
    document.getElementById('category-form').addEventListener('submit', async (e) => {
        e.preventDefault();

        const name = document.getElementById('category-name').value;
        const groupId = document.getElementById('category-group').value;
        const color = document.getElementById('category-color').value;
        const description = document.getElementById('category-description').value;

        try {
            await apiCall('/categories', {
                method: 'POST',
                body: JSON.stringify({
                    name,
                    group_id: groupId,
                    color,
                    description
                })
            });

            closeModal('category-modal');
            document.getElementById('category-form').reset();
            showToast('Category created successfully!');

            // Reload categories and budget
            await loadCategories();
            await loadBudgetView();
        } catch (error) {
            console.error('Failed to create category:', error);
        }
    });

    // Allocation form
    document.getElementById('allocation-form').addEventListener('submit', async (e) => {
        e.preventDefault();

        const categoryId = document.getElementById('allocation-category-id').value;
        const currentAmount = parseInt(document.getElementById('allocation-current-amount').value) || 0;
        const inputValue = document.getElementById('allocation-amount').value;
        const notes = document.getElementById('allocation-notes').value;
        const period = getCurrentPeriod();

        const result = parseAllocationInput(inputValue, currentAmount);

        if (!result.valid) {
            showToast(result.error, 'error');
            return;
        }

        try {
            await apiCall('/allocations', {
                method: 'POST',
                body: JSON.stringify({
                    category_id: categoryId,
                    amount: result.amountInCents,
                    period,
                    notes
                })
            });

            closeModal('allocation-modal');
            document.getElementById('allocation-form').reset();
            showToast('Budget allocated successfully!');

            // Reload budget view
            loadBudgetView();
        } catch (error) {
            console.error('Failed to create allocation:', error);
        }
    });

    // Import form
    document.getElementById('import-form').addEventListener('submit', async (e) => {
        e.preventDefault();

        const accountId = document.getElementById('import-account').value;
        const fileInput = document.getElementById('import-file');
        const file = fileInput.files[0];

        if (!file) {
            showToast('Please select a file', 'error');
            return;
        }

        const formData = new FormData();
        formData.append('account_id', accountId);
        formData.append('file', file);

        try {
            const button = e.target.querySelector('button[type="submit"]');
            button.disabled = true;
            button.textContent = 'Importing...';

            const response = await fetch('/api/transactions/import', {
                method: 'POST',
                body: formData
            });

            if (!response.ok) {
                const errorText = await response.text();
                throw new Error(errorText || 'Import failed');
            }

            const result = await response.json();

            button.disabled = false;
            button.textContent = 'Import Transactions';
            fileInput.value = '';

            showToast(`Imported ${result.imported_transactions} transactions (${result.skipped_duplicates} duplicates skipped)`);

            // Reload data
            await loadAccounts();
            await loadUncategorizedTransactions();
        } catch (error) {
            console.error('Failed to import:', error);
            showToast(error.message || 'Import failed', 'error');
            e.target.querySelector('button[type="submit"]').disabled = false;
            e.target.querySelector('button[type="submit"]').textContent = 'Import Transactions';
        }
    });

    // Categorize form
    document.getElementById('categorize-form').addEventListener('submit', async (e) => {
        e.preventDefault();

        const categoryId = document.getElementById('categorize-category').value;
        const selectedTransactions = window.selectedTransactions || [];

        if (selectedTransactions.length === 0) {
            showToast('No transactions selected', 'error');
            return;
        }

        try {
            await apiCall('/transactions/bulk-categorize', {
                method: 'POST',
                body: JSON.stringify({
                    transaction_ids: selectedTransactions,
                    category_id: categoryId || null
                })
            });

            closeModal('categorize-modal');
            document.getElementById('categorize-form').reset();
            window.selectedTransactions = [];
            showToast(`Categorized ${selectedTransactions.length} transaction(s)`);

            // Reload uncategorized transactions
            await loadUncategorizedTransactions();
        } catch (error) {
            console.error('Failed to categorize:', error);
            showToast('Failed to categorize transactions', 'error');
        }
    });

    // Initialize the app
    init();
});

// ============================================================================
// NEW SIDEBAR AND PANEL FUNCTIONS
// ============================================================================

// Render accounts in sidebar
async function renderAccountsSidebar() {
    const container = document.getElementById('sidebar-accounts-list');
    if (!container) return;

    if (accounts.length === 0) {
        container.innerHTML = '<p class="text-sm text-gray-500 dark:text-gray-400">No accounts yet</p>';
        return;
    }

    // Calculate total balance
    const totalBalance = accounts.reduce((sum, acc) => sum + acc.balance, 0);

    let html = `
        <div class="account-item cursor-pointer p-2 rounded hover:bg-gray-100 dark:hover:bg-gray-700 border-2 border-blue-500 dark:border-blue-400" onclick="loadAccountTransactionsPanel(null)">
            <div class="font-medium text-gray-900 dark:text-gray-100 text-sm">All Accounts</div>
            <div class="text-lg font-bold ${totalBalance >= 0 ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400'}">${formatCurrency(totalBalance)}</div>
        </div>
    `;

    accounts.forEach(account => {
        const isCreditCard = account.type === 'credit';
        const balanceClass = account.balance >= 0 ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400';

        // For credit cards, show "Owe $X.XX" instead of negative amount
        let balanceDisplay;
        if (isCreditCard && account.balance < 0) {
            balanceDisplay = `<span class="text-xs text-gray-500 dark:text-gray-400">Owe </span>${formatCurrency(Math.abs(account.balance))}`;
        } else {
            balanceDisplay = formatCurrency(account.balance);
        }

        html += `
            <div class="account-item cursor-pointer p-2 rounded hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors" onclick="loadAccountTransactionsPanel('${account.id}')">
                <div class="flex justify-between items-start">
                    <div class="font-medium text-gray-900 dark:text-gray-100 text-sm">${account.name}</div>
                    <span class="text-xs text-gray-500 dark:text-gray-400 capitalize">${account.type}</span>
                </div>
                <div class="text-sm font-semibold ${balanceClass}">${balanceDisplay}</div>
            </div>
        `;
    });

    container.innerHTML = html;
}

// Render uncategorized transactions in sidebar
async function renderUncategorizedTransactions() {
    const container = document.getElementById('sidebar-uncategorized-list');
    const countSpan = document.getElementById('uncategorized-count');
    if (!container) return;

    try {
        const uncategorized = await apiCall('/transactions?uncategorized=true') || [];

        // Filter to only show outflows (negative amounts) - inflows don't need categorization
        const outflows = uncategorized.filter(txn => txn.amount < 0);

        if (!outflows || outflows.length === 0) {
            container.innerHTML = '<p class="text-xs text-gray-500 dark:text-gray-400">All caught up!</p>';
            countSpan.textContent = '';
            return;
        }

        countSpan.textContent = `(${outflows.length})`;

        // Show first 5
        const toShow = outflows.slice(0, 5);
        let html = '';

        for (const txn of toShow) {
            const account = accounts.find(a => a.id === txn.account_id);
            const accountName = account ? account.name : 'Unknown';
            const desc = txn.description || 'Transaction';

            html += `
                <div class="text-xs bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded p-2">
                    <div class="font-medium text-gray-900 dark:text-gray-100 truncate" title="${desc}">${desc}</div>
                    <div class="flex items-center gap-1 mt-1">
                        <select class="text-xs border border-gray-300 dark:border-gray-600 rounded px-1 py-0.5 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 flex-1" onchange="quickCategorize('${txn.id}', this.value)">
                            <option value="">Category...</option>
                            ${categories.filter(c => !c.payment_for_account_id).map(cat =>
                                `<option value="${cat.id}">${cat.name}</option>`
                            ).join('')}
                        </select>
                        <span class="text-gray-600 dark:text-gray-400">${formatCurrency(txn.amount)}</span>
                    </div>
                </div>
            `;
        }

        if (outflows.length > 5) {
            html += `<p class="text-xs text-gray-500 dark:text-gray-400 mt-2">+${outflows.length - 5} more</p>`;
        }

        container.innerHTML = html;
    } catch (error) {
        console.error('Failed to load uncategorized transactions:', error);
        container.innerHTML = '<p class="text-xs text-red-500">Failed to load</p>';
    }
}

// Quick categorize from sidebar
async function quickCategorize(transactionId, categoryId) {
    if (!categoryId) return;

    try {
        // Fetch the existing transaction to get all fields
        const transaction = await apiCall(`/transactions/${transactionId}`);

        // Update with all fields, only changing the category
        await apiCall(`/transactions/${transactionId}`, {
            method: 'PUT',
            body: JSON.stringify({
                account_id: transaction.account_id,
                category_id: categoryId,
                amount: transaction.amount,
                description: transaction.description,
                date: transaction.date
            })
        });

        showToast('Transaction categorized!');

        // Reload sidebar sections
        await renderUncategorizedTransactions();
        await renderRecentTransactions();
        await loadBudgetView(); // Reload budget to update spending
    } catch (error) {
        console.error('Failed to categorize:', error);
        showToast('Failed to categorize transaction', 'error');
    }
}

// Make quickCategorize available globally
window.quickCategorize = quickCategorize;

// Render recent transactions in sidebar
async function renderRecentTransactions() {
    const container = document.getElementById('sidebar-recent-list');
    if (!container) return;

    try {
        // Get all transactions and sort by date
        const allTransactions = await apiCall('/transactions') || [];
        const recent = allTransactions
            .filter(t => t.category_id) // Only categorized
            .sort((a, b) => new Date(b.date) - new Date(a.date))
            .slice(0, 10);

        if (recent.length === 0) {
            container.innerHTML = '<p class="text-xs text-gray-500 dark:text-gray-400">No recent activity</p>';
            return;
        }

        let html = '';
        for (const txn of recent) {
            const category = categories.find(c => c.id === txn.category_id);
            const desc = txn.description || 'Transaction';
            const amountClass = txn.amount >= 0 ? 'text-green-600 dark:text-green-400' : 'text-gray-700 dark:text-gray-300';

            html += `
                <div class="text-xs border-b border-gray-100 dark:border-gray-700 pb-1.5 mb-1.5 last:border-b-0">
                    <div class="flex justify-between items-start">
                        <span class="font-medium text-gray-900 dark:text-gray-100 truncate flex-1" title="${desc}">${desc}</span>
                        <span class="${amountClass} font-semibold ml-2">${formatCurrency(txn.amount)}</span>
                    </div>
                    <div class="flex items-center gap-1 mt-0.5">
                        ${category ? `<span class="w-2 h-2 rounded-full" style="background-color: ${category.color}"></span>` : ''}
                        <span class="text-gray-500 dark:text-gray-400">${category ? category.name : 'Uncategorized'}</span>
                    </div>
                </div>
            `;
        }

        container.innerHTML = html;
    } catch (error) {
        console.error('Failed to load recent transactions:', error);
        container.innerHTML = '<p class="text-xs text-red-500">Failed to load</p>';
    }
}

// Load sidebar data
async function loadSidebar() {
    await renderAccountsSidebar();
    await renderUncategorizedTransactions();
    await renderRecentTransactions();
}

// Open transaction panel for account
async function loadAccountTransactionsPanel(accountId) {
    const panel = document.getElementById('transaction-panel');
    const backdrop = document.getElementById('transaction-panel-backdrop');
    const content = document.getElementById('transaction-panel-content');
    const title = document.getElementById('transaction-panel-title');
    const subtitle = document.getElementById('transaction-panel-subtitle');

    if (!panel || !backdrop || !content) return;

    try {
        let transactions;
        let accountName = 'All Accounts';

        if (accountId) {
            // Load specific account transactions
            transactions = await apiCall(`/accounts/${accountId}/transactions`);
            const account = accounts.find(a => a.id === accountId);
            if (account) {
                accountName = account.name;
                subtitle.textContent = `Balance: ${formatCurrency(account.balance)}`;
            }
        } else {
            // Load all transactions
            transactions = await apiCall('/transactions');
            const totalBalance = accounts.reduce((sum, acc) => sum + acc.balance, 0);
            subtitle.textContent = `Total Balance: ${formatCurrency(totalBalance)}`;
        }

        title.textContent = accountName;

        // Sort by date descending
        transactions.sort((a, b) => new Date(b.date) - new Date(a.date));

        // Render transactions
        if (transactions.length === 0) {
            content.innerHTML = '<p class="text-gray-500 dark:text-gray-400">No transactions yet</p>';
        } else {
            let html = '<div class="space-y-2">';

            transactions.forEach(txn => {
                const account = accounts.find(a => a.id === txn.account_id);
                const category = categories.find(c => c.id === txn.category_id);
                const desc = txn.description || 'Transaction';
                const amountClass = txn.amount >= 0 ? 'text-green-600 dark:text-green-400' : 'text-gray-700 dark:text-gray-300';

                // Check if it's a transfer
                const isTransfer = txn.type === 'transfer';
                let displayText = desc;
                if (isTransfer && txn.transfer_to_account_id) {
                    const toAccount = accounts.find(a => a.id === txn.transfer_to_account_id);
                    if (toAccount) {
                        displayText = txn.amount < 0
                            ? `Transfer to ${toAccount.name}`
                            : `Transfer from ${toAccount.name}`;
                    }
                }

                html += `
                    <div class="border border-gray-200 dark:border-gray-700 rounded-lg p-3 hover:bg-gray-50 dark:hover:bg-gray-700/50">
                        <div class="flex justify-between items-start mb-1">
                            <div class="font-medium text-gray-900 dark:text-gray-100">${displayText}</div>
                            <div class="text-lg font-semibold ${amountClass}">${formatCurrency(txn.amount)}</div>
                        </div>
                        <div class="text-sm text-gray-600 dark:text-gray-400 space-y-0.5">
                            <div>${formatDate(txn.date)}</div>
                            ${accountId ? '' : `<div>Account: ${account ? account.name : 'Unknown'}</div>`}
                            ${category ? `
                                <div class="flex items-center gap-1">
                                    <span class="w-3 h-3 rounded-full" style="background-color: ${category.color}"></span>
                                    <span>${category.name}</span>
                                </div>
                            ` : '<div class="text-yellow-600 dark:text-yellow-400">Uncategorized</div>'}
                        </div>
                    </div>
                `;
            });

            html += '</div>';
            content.innerHTML = html;
        }

        // Show panel
        backdrop.classList.remove('hidden');
        panel.classList.remove('translate-x-full');
    } catch (error) {
        console.error('Failed to load account transactions:', error);
        showToast('Failed to load transactions', 'error');
    }
}

// Close transaction panel
function closeTransactionPanel() {
    const panel = document.getElementById('transaction-panel');
    const backdrop = document.getElementById('transaction-panel-backdrop');

    if (panel && backdrop) {
        panel.classList.add('translate-x-full');
        backdrop.classList.add('hidden');
    }
}

// Show import view (modal)
function showImportView() {
    const modal = document.getElementById('import-modal');
    if (modal) {
        modal.classList.add('active');
        // Populate account dropdown
        const select = document.getElementById('import-account');
        if (select) {
            select.innerHTML = '<option value="">Choose account to import into...</option>';
            accounts.forEach(account => {
                const option = document.createElement('option');
                option.value = account.id;
                option.textContent = `${account.name} (${account.type})`;
                select.appendChild(option);
            });
        }
    }
}

// Close import view
function closeImportView() {
    const modal = document.getElementById('import-modal');
    if (modal) {
        modal.classList.remove('active');
    }
}

// Make functions globally available
window.loadAccountTransactionsPanel = loadAccountTransactionsPanel;
window.closeTransactionPanel = closeTransactionPanel;
window.showImportView = showImportView;
window.closeImportView = closeImportView;

// ============================================================================
// END NEW SIDEBAR AND PANEL FUNCTIONS
// ============================================================================

// Initialize the app
async function init() {
    try {
        await loadAccounts();
        await loadCategories();
        await loadBudgetView();
        await loadSidebar(); // Load sidebar data

        // Show helpful message if starting fresh
        if (accounts.length === 0 && categories.length === 0) {
            showToast('Welcome! Start by creating an account and some categories.', 'success');
        }
    } catch (error) {
        console.error('Failed to initialize app:', error);
    }
}
