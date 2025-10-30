// Global state
let currentMonth = new Date();
let accounts = [];
let categories = [];
let categoryGroups = [];
let transactions = [];
let allocations = [];

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

// View management
function showView(viewName) {
    // Update navigation
    document.querySelectorAll('.nav-item').forEach(item => {
        item.classList.remove('active');
        if (item.dataset.view === viewName) {
            item.classList.add('active');
        }
    });

    // Hide all views
    document.querySelectorAll('.view').forEach(view => {
        view.classList.add('hidden');
    });

    // Show selected view
    const viewElement = document.getElementById(`${viewName}-view`);
    if (viewElement) {
        viewElement.classList.remove('hidden');
    }

    // Load data for the view
    switch(viewName) {
        case 'budget':
            loadBudgetView();
            break;
        case 'accounts':
            loadAccountsView();
            break;
        case 'transactions':
            loadTransactionsView();
            break;
        case 'import':
            loadImportView();
            break;
        case 'categories':
            loadCategoriesView();
            break;
    }
}

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
        readyToAssignEl.textContent = formatCurrency(readyToAssign);
        if (readyToAssign < 0) {
            readyToAssignEl.className = 'text-3xl font-bold text-red-600';
        } else {
            readyToAssignEl.className = 'text-3xl font-bold text-blue-600';
        }

        const budgetCategories = document.getElementById('budget-categories');

        if (categories.length === 0) {
            budgetCategories.innerHTML = `
                <div class="text-center py-12">
                    <p class="text-gray-500 mb-4">No expense categories yet.</p>
                    <button onclick="showView('categories')" class="btn-primary">Create Your First Category</button>
                </div>
            `;
            return;
        }

        // Render groups and ungrouped categories
        budgetCategories.innerHTML = renderBudgetWithGroups(summary);

        // Initialize drag-and-drop after rendering
        initializeBudgetDragDrop();
    } catch (error) {
        console.error('Failed to load budget view:', error);
    }
}

function renderBudgetWithGroups(summary) {
    let html = '';

    // Sort groups by display order
    const sortedGroups = [...categoryGroups].sort((a, b) => a.display_order - b.display_order);

    // Render each group (including empty ones)
    for (const group of sortedGroups) {
        const groupCategories = categories.filter(c => c.group_id === group.id);
        html += renderGroupSection(group, groupCategories, summary);
    }

    // Always render ungrouped section
    const ungroupedCategories = categories.filter(c => !c.group_id);
    html += renderUngroupedSection(ungroupedCategories, summary);

    return html;
}

function renderGroupSection(group, groupCategories, summary) {
    const categoriesHtml = groupCategories.length > 0
        ? groupCategories.map(cat => renderBudgetCategory(cat, summary)).join('')
        : '<div class="text-gray-400 text-sm p-4 border-2 border-dashed border-gray-200 rounded text-center">Drag categories here</div>';

    return `
        <div class="budget-group mb-4" data-group-id="${group.id}">
            <div class="flex items-center gap-2 mb-2 p-2 bg-gray-100 rounded cursor-move hover:bg-gray-200 transition">
                <span class="drag-handle text-gray-400">⋮⋮</span>
                <h3 class="text-lg font-semibold text-gray-700 flex-1">${group.name}</h3>
                <button onclick="deleteGroup('${group.id}')" class="text-xs text-red-600 hover:text-red-800">Delete</button>
            </div>
            <div class="group-categories space-y-2 min-h-[60px]" data-group-id="${group.id}">
                ${categoriesHtml}
            </div>
        </div>
    `;
}

function renderUngroupedSection(ungroupedCategories, summary) {
    const categoriesHtml = ungroupedCategories.length > 0
        ? ungroupedCategories.map(cat => renderBudgetCategory(cat, summary)).join('')
        : '<div class="text-gray-400 text-sm p-4 border-2 border-dashed border-gray-200 rounded text-center">Drag categories here to ungroup</div>';

    return `
        <div class="budget-group mb-4" data-group-id="ungrouped">
            <h3 class="text-lg font-semibold text-gray-500 mb-2 p-2">Ungrouped</h3>
            <div class="group-categories space-y-2 min-h-[60px]" data-group-id="ungrouped">
                ${categoriesHtml}
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

    const allocatedDisplay = isPaymentCategory
        ? `<div class="font-semibold" title="Auto-allocated">${formatCurrency(allocated)}</div>`
        : `<div class="font-semibold cursor-pointer hover:bg-blue-50 rounded px-2 py-1 -mx-2 -my-1 no-drag"
                onclick="event.stopPropagation(); startInlineEdit('${category.id}', '${category.name.replace(/'/g, "\\'")}', ${allocated})"
                title="Click to edit">${formatCurrency(allocated)}</div>`;

    const underfundedWarning = isUnderfunded
        ? `<div class="mt-2 p-2 bg-red-50 border border-red-200 rounded text-sm">
            <span class="text-red-600 font-semibold">⚠️ Underfunded - Need ${formatCurrency(summaryItem.underfunded)} more</span>
        </div>` : '';

    return `
        <div class="budget-category border border-gray-200 rounded-lg p-4 bg-white cursor-move ${isPaymentCategory ? 'bg-orange-50' : ''}"
             data-category-id="${category.id}">
            <div class="flex justify-between items-center">
                <div class="flex items-center gap-3 flex-1">
                    <span class="text-gray-400 text-xs">⋮⋮</span>
                    <div class="w-3 h-3 rounded-full flex-shrink-0" style="background-color: ${category.color || '#3b82f6'}"></div>
                    <div class="flex-1">
                        <div class="font-semibold text-gray-800">${category.name}</div>
                    </div>
                </div>
                <div class="flex gap-6 items-center">
                    <div class="text-right">
                        <div class="text-xs text-gray-500">Allocated</div>
                        ${allocatedDisplay}
                    </div>
                    <div class="text-right">
                        <div class="text-xs text-gray-500">Spent</div>
                        <div class="font-semibold">${formatCurrency(spent)}</div>
                    </div>
                    <div class="text-right min-w-[100px]">
                        <div class="text-xs text-gray-500">Available</div>
                        <div class="font-bold ${availableClass}">${formatCurrency(available)}</div>
                    </div>
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
                onEnd: async function(evt) {
                    const categoryId = evt.item.dataset.categoryId;
                    const newGroupId = evt.to.dataset.groupId;
                    await updateCategoryGroup(categoryId, newGroupId === 'ungrouped' ? null : newGroupId);
                }
            });
        });
    }
}

async function updateGroupOrder() {
    const groups = [...document.querySelectorAll('.budget-group[data-group-id]:not([data-group-id="ungrouped"])')];
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
        if (groupId) {
            await apiCall('/category-groups/assign', {
                method: 'POST',
                body: JSON.stringify({ category_id: categoryId, group_id: groupId })
            });
        } else {
            await apiCall(`/category-groups/unassign/${categoryId}`, { method: 'POST' });
        }
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
    if (!confirm('Delete this group? Categories will be moved to Ungrouped.')) return;

    try {
        await apiCall(`/category-groups/${groupId}`, { method: 'DELETE' });
        await loadCategoryGroups();
        loadBudgetView();
        showToast('Group deleted successfully!');
    } catch (error) {
        console.error('Failed to delete group:', error);
    }
}

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
                    <p class="text-gray-500 mb-4">No accounts yet. Create one to start tracking your money!</p>
                    <button onclick="showAddAccountModal()" class="btn-primary">Create Your First Account</button>
                </div>
            `;
            return;
        }

        accountsList.innerHTML = accounts.map(account => {
            const balanceClass = account.balance >= 0 ? 'text-green-600' : 'text-red-600';
            return `
                <div class="border border-gray-200 rounded-lg p-4 hover:shadow-md transition-shadow">
                    <div class="flex justify-between items-center">
                        <div>
                            <div class="font-semibold text-gray-800">${account.name}</div>
                            <div class="text-sm text-gray-500 capitalize">${account.type}</div>
                        </div>
                        <div class="text-right">
                            <div class="text-xl font-bold ${balanceClass}">${formatCurrency(account.balance)}</div>
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
                    <p class="text-gray-500 mb-4">No transactions yet.</p>
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
                <div class="border border-gray-200 rounded-lg p-4 hover:shadow-md transition-shadow">
                    <div class="flex justify-between items-center">
                        <div class="flex-1">
                            <div class="flex items-center gap-2">
                                ${category ? `<div class="w-2 h-2 rounded-full" style="background-color: ${category.color || '#gray'}"></div>` : ''}
                                <div class="font-semibold text-gray-800">${transaction.description || 'Transaction'}</div>
                            </div>
                            <div class="text-sm text-gray-500 mt-1">
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
            categoriesList.innerHTML = '<div class="text-gray-500 text-center py-4">No categories yet.</div>';
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
        <div class="border border-gray-200 rounded-lg p-4 hover:shadow-md transition-shadow">
            <div class="flex items-center gap-3">
                <div class="w-4 h-4 rounded-full flex-shrink-0" style="background-color: ${category.color || '#3b82f6'}"></div>
                <div class="flex-1">
                    <div class="font-semibold text-gray-800">${category.name}</div>
                    ${category.description ? `<div class="text-sm text-gray-500">${category.description}</div>` : ''}
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
        showView('accounts');
        return;
    }

    if (categories.length === 0) {
        showToast('Please create a category first', 'error');
        showView('categories');
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
    showModal('category-modal');
}

function showAllocateModal(categoryId, categoryName, currentAmount = 0) {
    document.getElementById('allocation-category-id').value = categoryId;
    document.getElementById('allocation-category-name').textContent = categoryName;
    document.getElementById('allocation-amount').value = (currentAmount / 100).toFixed(2);
    document.getElementById('allocation-notes').value = '';
    showModal('allocation-modal');
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
    input.type = 'number';
    input.step = '0.01';
    input.min = '0';
    input.value = (currentAmount / 100).toFixed(2);
    input.className = 'w-24 border border-blue-500 rounded px-2 py-1 text-center font-semibold focus:outline-none focus:ring-2 focus:ring-blue-500';

    // Replace content with input
    clickedElement.innerHTML = '';
    clickedElement.appendChild(input);
    input.focus();
    input.select();

    // Function to save the allocation
    const saveAllocation = async () => {
        const newAmount = parseFloat(input.value);

        if (isNaN(newAmount) || newAmount < 0) {
            showToast('Please enter a valid amount', 'error');
            clickedElement.innerHTML = originalContent;
            return;
        }

        const amountInCents = Math.round(newAmount * 100);

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

// Load uncategorized transactions
async function loadUncategorizedTransactions() {
    try {
        const transactions = await apiCall('/transactions?uncategorized=true');

        const listContainer = document.getElementById('uncategorized-list');

        if (transactions.length === 0) {
            listContainer.innerHTML = '<p class="text-gray-500 text-center py-4">No uncategorized transactions</p>';
            return;
        }

        listContainer.innerHTML = `
            <div class="mb-3 flex gap-2">
                <button onclick="selectAllUncategorized()" class="btn-secondary text-sm">Select All</button>
                <button onclick="showCategorizeModal()" class="btn-primary text-sm">Categorize Selected</button>
            </div>
            ${transactions.map(txn => {
                const account = accounts.find(a => a.id === txn.account_id);
                const amountClass = txn.amount >= 0 ? 'text-green-600' : 'text-red-600';
                return `
                    <div class="flex items-center gap-3 p-3 bg-white rounded-lg border border-gray-200 hover:border-blue-300 transition">
                        <input type="checkbox" class="uncategorized-checkbox" data-transaction-id="${txn.id}">
                        <div class="flex-1 min-w-0">
                            <div class="flex justify-between items-start gap-2">
                                <div class="flex-1 min-w-0">
                                    <div class="font-medium text-gray-800 truncate">${txn.description || 'No description'}</div>
                                    <div class="text-xs text-gray-500">${account ? account.name : 'Unknown'} • ${new Date(txn.date).toLocaleDateString()}</div>
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
    // Add listener for transaction type change to update category requirement
    document.getElementById('transaction-type').addEventListener('change', function() {
        const categorySelect = document.getElementById('transaction-category');
        const categoryIndicator = document.getElementById('category-required-indicator');

        if (this.value === 'inflow') {
            // Income: category is optional
            categorySelect.removeAttribute('required');
            categoryIndicator.textContent = '';
        } else {
            // Expense: category is required
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
        const amount = parseFloat(document.getElementById('transaction-amount').value);
        const type = document.getElementById('transaction-type').value;
        const date = document.getElementById('transaction-date').value;
        const description = document.getElementById('transaction-description').value;

        if (!accountId) {
            showToast('Please select an account', 'error');
            return;
        }

        // Category is required for outflow but optional for inflow
        if (type === 'outflow' && !categoryId) {
            showToast('Please select a category for expenses', 'error');
            return;
        }

        // Convert amount to cents, negative for outflow
        const amountInCents = Math.round((type === 'outflow' ? -amount : amount) * 100);

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

            // Reload views
            loadBudgetView();
            loadAccountsView();
            loadTransactionsView();
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

            // Reload views (including budget to show payment category updates)
            loadBudgetView();
            loadAccountsView();
            loadTransactionsView();
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

            // Reload accounts
            await loadAccounts();
            loadAccountsView();
        } catch (error) {
            console.error('Failed to create account:', error);
        }
    });

    // Category form
    document.getElementById('category-form').addEventListener('submit', async (e) => {
        e.preventDefault();

        const name = document.getElementById('category-name').value;
        const color = document.getElementById('category-color').value;
        const description = document.getElementById('category-description').value;

        try {
            await apiCall('/categories', {
                method: 'POST',
                body: JSON.stringify({
                    name,
                    color,
                    description
                })
            });

            closeModal('category-modal');
            document.getElementById('category-form').reset();
            showToast('Category created successfully!');

            // Reload categories
            await loadCategories();
            loadCategoriesView();
            loadBudgetView();
        } catch (error) {
            console.error('Failed to create category:', error);
        }
    });

    // Allocation form
    document.getElementById('allocation-form').addEventListener('submit', async (e) => {
        e.preventDefault();

        const categoryId = document.getElementById('allocation-category-id').value;
        const amount = parseFloat(document.getElementById('allocation-amount').value);
        const notes = document.getElementById('allocation-notes').value;
        const period = getCurrentPeriod();

        try {
            await apiCall('/allocations', {
                method: 'POST',
                body: JSON.stringify({
                    category_id: categoryId,
                    amount: Math.round(amount * 100),
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

// Initialize the app
async function init() {
    try {
        await loadAccounts();
        await loadCategories();
        await loadBudgetView();

        // Show helpful message if starting fresh
        if (accounts.length === 0 && categories.length === 0) {
            showToast('Welcome! Start by creating an account and some categories.', 'success');
        }
    } catch (error) {
        console.error('Failed to initialize app:', error);
    }
}
