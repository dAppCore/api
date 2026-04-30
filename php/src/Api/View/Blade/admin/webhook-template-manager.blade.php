<div class="space-y-6">
    {{-- Header --}}
    <div class="flex items-center justify-between">
        <div>
            <core:heading size="xl">Webhook templates</core:heading>
            <core:subheading>
                Reusable templates for customising webhook payload shapes.
            </core:subheading>
        </div>

        <core:button wire:click="create" variant="primary" icon="plus">
            Create template
        </core:button>
    </div>

    {{-- Filters and Search --}}
    <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div class="flex items-center gap-2">
            <core:select wire:model.live="filter" class="w-40">
                <option value="all">All templates</option>
                <option value="custom">Custom only</option>
                <option value="builtin">Built-in only</option>
                <option value="active">Active</option>
                <option value="inactive">Inactive</option>
            </core:select>
        </div>

        <div class="w-full sm:w-64">
            <core:input
                wire:model.live.debounce.300ms="search"
                type="search"
                placeholder="Search templates..."
                icon="magnifying-glass"
            />
        </div>
    </div>

    {{-- Templates List --}}
    <flux:card>
        @if($this->templates->isEmpty())
            <div class="py-12 text-center">
                <core:icon name="document-text" class="mx-auto h-12 w-12 text-zinc-400" />
                <core:heading size="lg" class="mt-4">No templates found</core:heading>
                <core:subheading class="mt-2">
                    @if($search)
                        No templates match your search criteria.
                    @else
                        Create a custom template or use one of the built-in templates.
                    @endif
                </core:subheading>
            </div>
        @else
            <div class="divide-y divide-zinc-200 dark:divide-zinc-700">
                @foreach($this->templates as $template)
                    <div class="flex items-center justify-between p-4 hover:bg-zinc-50 dark:hover:bg-zinc-800/50">
                        <div class="flex items-center gap-4">
                            {{-- Icon --}}
                            <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-zinc-100 dark:bg-zinc-800">
                                <core:icon name="{{ $template->type_icon }}" class="h-5 w-5 text-zinc-600 dark:text-zinc-400" />
                            </div>

                            {{-- Info --}}
                            <div>
                                <div class="flex items-center gap-2">
                                    <span class="font-medium text-zinc-900 dark:text-white">{{ $template->name }}</span>
                                    @if($template->isBuiltin())
                                        <flux:badge color="purple" size="sm">Built-in</flux:badge>
                                    @endif
                                    @if($template->is_default)
                                        <flux:badge color="green" size="sm">Default</flux:badge>
                                    @endif
                                    @if(!$template->is_active)
                                        <flux:badge color="zinc" size="sm">Inactive</flux:badge>
                                    @endif
                                </div>
                                <p class="mt-0.5 text-sm text-zinc-500 dark:text-zinc-400">
                                    {{ $template->description ?? 'No description' }}
                                </p>
                                <p class="mt-0.5 text-xs text-zinc-400 dark:text-zinc-500">
                                    Format: {{ $template->format->label() }}
                                </p>
                            </div>
                        </div>

                        {{-- Actions --}}
                        <div class="flex items-center gap-2">
                            <core:button wire:click="edit('{{ $template->uuid }}')" variant="ghost" size="sm" icon="pencil">
                                Edit
                            </core:button>

                            <flux:dropdown>
                                <flux:button variant="ghost" size="sm" icon="ellipsis-vertical" />

                                <flux:menu>
                                    @if(!$template->is_default)
                                        <flux:menu.item wire:click="setDefault('{{ $template->uuid }}')" icon="star">
                                            Set as default
                                        </flux:menu.item>
                                    @endif

                                    <flux:menu.item wire:click="duplicate('{{ $template->uuid }}')" icon="document-duplicate">
                                        Duplicate
                                    </flux:menu.item>

                                    <flux:menu.item wire:click="toggleActive('{{ $template->uuid }}')" icon="{{ $template->is_active ? 'pause' : 'play' }}">
                                        {{ $template->is_active ? 'Disable' : 'Enable' }}
                                    </flux:menu.item>

                                    @if(!$template->isBuiltin())
                                        <flux:menu.separator />
                                        <flux:menu.item wire:click="confirmDelete('{{ $template->uuid }}')" icon="trash" variant="danger">
                                            Delete
                                        </flux:menu.item>
                                    @endif
                                </flux:menu>
                            </flux:dropdown>
                        </div>
                    </div>
                @endforeach
            </div>

            {{-- Pagination --}}
            @if($this->templates->hasPages())
                <div class="border-t border-zinc-200 px-4 py-3 dark:border-zinc-700">
                    {{ $this->templates->links() }}
                </div>
            @endif
        @endif
    </flux:card>

    {{-- Delete Confirmation Modal --}}
    @if($deletingId)
        <flux:modal wire:model="deletingId" class="max-w-md">
            <flux:heading size="lg">Delete template</flux:heading>
            <flux:subheading class="mt-2">
                Are you sure you want to delete this template? This action cannot be undone.
            </flux:subheading>

            <div class="mt-6 flex justify-end gap-3">
                <core:button wire:click="cancelDelete" variant="ghost">
                    Cancel
                </core:button>
                <core:button wire:click="delete" variant="danger">
                    Delete
                </core:button>
            </div>
        </flux:modal>
    @endif

    {{-- Editor Modal --}}
    @if($showEditor)
        <flux:modal wire:model="showEditor" class="max-w-5xl">
            <form wire:submit="save" class="space-y-6">
                <flux:heading size="lg">
                    {{ $editingId ? 'Edit template' : 'Create template' }}
                </flux:heading>

                <div class="grid grid-cols-1 gap-6 lg:grid-cols-3">
                    {{-- Main editor (2/3 width) --}}
                    <div class="lg:col-span-2 space-y-4">
                        {{-- Name --}}
                        <div>
                            <core:label for="name">Template name</core:label>
                            <core:input
                                wire:model="name"
                                id="name"
                                placeholder="e.g. Slack notification"
                                class="mt-1"
                            />
                            @error('name')
                                <core:text class="mt-1 text-sm text-red-600">{{ $message }}</core:text>
                            @enderror
                        </div>

                        {{-- Description --}}
                        <div>
                            <core:label for="description">Description</core:label>
                            <core:input
                                wire:model="description"
                                id="description"
                                placeholder="Brief description of this template"
                                class="mt-1"
                            />
                        </div>

                        {{-- Format selector --}}
                        <div>
                            <core:label for="format">Template format</core:label>
                            <core:select wire:model.live="format" id="format" class="mt-1">
                                @foreach($this->templateFormats as $value => $label)
                                    <option value="{{ $value }}">{{ $label }}</option>
                                @endforeach
                            </core:select>
                            <core:text class="mt-1 text-xs text-zinc-500">
                                {{ $this->templateFormatDescriptions[$format] ?? '' }}
                            </core:text>
                        </div>

                        {{-- Template textarea --}}
                        <div
                            x-data="{
                                insertAtCursor(text) {
                                    const textarea = this.$refs.templateEditor;
                                    const start = textarea.selectionStart;
                                    const end = textarea.selectionEnd;
                                    const value = textarea.value;
                                    textarea.value = value.substring(0, start) + text + value.substring(end);
                                    textarea.selectionStart = textarea.selectionEnd = start + text.length;
                                    textarea.focus();
                                    $wire.set('template', textarea.value);
                                }
                            }"
                            @insert-variable.window="insertAtCursor($event.detail.variable)"
                        >
                            <core:label for="template">Template content</core:label>
                            <textarea
                                x-ref="templateEditor"
                                wire:model.blur="template"
                                id="template"
                                rows="12"
                                class="mt-1 w-full rounded-lg border border-zinc-200 bg-zinc-50 p-3 font-mono text-sm dark:border-zinc-700 dark:bg-zinc-800"
                                placeholder='{"event": "{{event.type}}", "data": {{data | json}}}'
                            ></textarea>
                            @error('template')
                                <core:text class="mt-1 text-sm text-red-600">{{ $message }}</core:text>
                            @enderror
                        </div>

                        {{-- Template errors --}}
                        @if($templateErrors)
                            <div class="rounded-lg border border-red-200 bg-red-50 p-3 dark:border-red-800 dark:bg-red-900/20">
                                <p class="font-medium text-red-800 dark:text-red-200">Template errors:</p>
                                <ul class="mt-1 list-inside list-disc text-sm text-red-700 dark:text-red-300">
                                    @foreach($templateErrors as $error)
                                        <li>{{ $error }}</li>
                                    @endforeach
                                </ul>
                            </div>
                        @endif

                        {{-- Action buttons --}}
                        <div class="flex flex-wrap items-center gap-2">
                            <core:button type="button" wire:click="previewTemplate" variant="ghost" size="sm">
                                Preview output
                            </core:button>
                            <core:button type="button" wire:click="validateTemplate" variant="ghost" size="sm">
                                Validate
                            </core:button>
                            <core:button type="button" wire:click="resetTemplate" variant="ghost" size="sm">
                                Reset to default
                            </core:button>
                        </div>

                        {{-- Preview output --}}
                        @if($templatePreview)
                            <div class="rounded-lg border border-green-200 bg-green-50 p-3 dark:border-green-800 dark:bg-green-900/20">
                                <p class="mb-2 font-medium text-green-800 dark:text-green-200">Preview output:</p>
                                <pre class="overflow-x-auto rounded bg-white p-2 text-xs text-zinc-800 dark:bg-zinc-900 dark:text-zinc-200">{{ json_encode($templatePreview, JSON_PRETTY_PRINT | JSON_UNESCAPED_SLASHES) }}</pre>
                            </div>
                        @endif

                        {{-- Options --}}
                        <div class="flex items-center gap-6">
                            <label class="flex items-center gap-2">
                                <core:checkbox wire:model="isDefault" />
                                <span class="text-sm">Set as default template</span>
                            </label>
                            <label class="flex items-center gap-2">
                                <core:checkbox wire:model="isActive" />
                                <span class="text-sm">Template is active</span>
                            </label>
                        </div>
                    </div>

                    {{-- Sidebar (1/3 width) --}}
                    <div class="space-y-4">
                        {{-- Load from builtin --}}
                        <div>
                            <core:label>Load from built-in template</core:label>
                            <div class="mt-2 space-y-1">
                                @foreach($this->builtinTemplates as $type => $info)
                                    <button
                                        type="button"
                                        wire:click="loadBuiltinTemplate('{{ $type }}')"
                                        class="flex w-full items-center gap-2 rounded-lg border border-zinc-200 p-2 text-left text-sm transition hover:bg-zinc-50 dark:border-zinc-700 dark:hover:bg-zinc-700/50"
                                    >
                                        <span class="font-medium">{{ $info['name'] }}</span>
                                    </button>
                                @endforeach
                            </div>
                        </div>

                        {{-- Available variables --}}
                        <div>
                            <core:label>Available variables</core:label>
                            <div class="mt-2 max-h-48 space-y-1 overflow-y-auto rounded-lg border border-zinc-200 bg-zinc-50 p-2 dark:border-zinc-700 dark:bg-zinc-800">
                                @foreach($this->availableVariables as $variable => $info)
                                    <button
                                        type="button"
                                        wire:click="insertVariable('{{ $variable }}')"
                                        class="flex w-full items-start gap-2 rounded p-1.5 text-left text-xs transition hover:bg-zinc-200 dark:hover:bg-zinc-700"
                                        title="{{ $info['description'] }}"
                                    >
                                        <code class="shrink-0 rounded bg-zinc-200 px-1 font-mono text-zinc-700 dark:bg-zinc-600 dark:text-zinc-200">{{ '{{' . $variable . '}}' }}</code>
                                        <span class="text-zinc-500 dark:text-zinc-400">{{ $info['type'] }}</span>
                                    </button>
                                @endforeach
                            </div>
                            <p class="mt-1 text-xs text-zinc-500">Click to insert at cursor position.</p>
                        </div>

                        {{-- Available filters --}}
                        <div>
                            <core:label>Available filters</core:label>
                            <div class="mt-2 space-y-1 rounded-lg border border-zinc-200 bg-zinc-50 p-2 dark:border-zinc-700 dark:bg-zinc-800">
                                @foreach($this->availableFilters as $filter => $description)
                                    <div class="text-xs">
                                        <code class="rounded bg-zinc-200 px-1 font-mono text-zinc-700 dark:bg-zinc-600 dark:text-zinc-200">| {{ $filter }}</code>
                                        <span class="text-zinc-500 dark:text-zinc-400">{{ $description }}</span>
                                    </div>
                                @endforeach
                            </div>
                        </div>

                        {{-- Syntax help --}}
                        <div class="rounded-lg border border-zinc-200 bg-zinc-50 p-3 dark:border-zinc-700 dark:bg-zinc-800">
                            <p class="mb-2 text-xs font-medium text-zinc-700 dark:text-zinc-300">Syntax reference</p>
                            <div class="space-y-1 text-xs text-zinc-600 dark:text-zinc-400">
                                <p><code class="rounded bg-zinc-200 px-1 dark:bg-zinc-600">@{{ '{{variable}}' }}</code> - Simple value</p>
                                <p><code class="rounded bg-zinc-200 px-1 dark:bg-zinc-600">@{{ '{{data.nested}}' }}</code> - Nested value</p>
                                <p><code class="rounded bg-zinc-200 px-1 dark:bg-zinc-600">@{{ '{{value | filter}}' }}</code> - With filter</p>
                                @if($format === 'mustache')
                                    <p><code class="rounded bg-zinc-200 px-1 dark:bg-zinc-600">@{{ '{{#if var}}...{{/if}}' }}</code> - Conditional</p>
                                    <p><code class="rounded bg-zinc-200 px-1 dark:bg-zinc-600">@{{ '{{#each arr}}...{{/each}}' }}</code> - Loop</p>
                                @endif
                            </div>
                        </div>
                    </div>
                </div>

                {{-- Footer actions --}}
                <div class="flex justify-end gap-3 border-t border-zinc-200 pt-4 dark:border-zinc-700">
                    <core:button type="button" wire:click="closeEditor" variant="ghost">
                        Cancel
                    </core:button>
                    <core:button type="submit" variant="primary">
                        {{ $editingId ? 'Update template' : 'Create template' }}
                    </core:button>
                </div>
            </form>
        </flux:modal>
    @endif
</div>
