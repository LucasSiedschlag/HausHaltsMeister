<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { Button } from '@shared/components/ui/button'
import CrudTableCard from '@shared/components/CrudTableCard.vue'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@shared/components/ui/dropdown-menu'
import { Input } from '@shared/components/ui/input'
import { Label } from '@shared/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@shared/components/ui/select'
import { Sheet, SheetContent, SheetDescription, SheetFooter, SheetHeader, SheetTitle } from '@shared/components/ui/sheet'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@shared/components/ui/tooltip'
import { MoreHorizontal, Plus } from 'lucide-vue-next'
import { push } from 'notivue'
import { useAccounts, type Account, type AccountType } from '#layers/accounts/composables/useAccounts'
import { useLedgerContext } from '@shared/composables/useLedgerContext'
import { useHeaderAction } from '@shared/composables/useHeaderAction'
import { accountTypeSchema, nameSchema, useInlineValidation, getInputClass } from '@shared/validators'
import { isApiError } from '@shared/utils/api-error'
import ConfirmDialog from '@shared/components/ConfirmDialog.vue'

const { t } = useI18n()
const ledgerContext = useLedgerContext()
const canEdit = computed(() => ledgerContext.hasRole('editor'))

const { accounts, loading, error, fetchAccounts, createAccount, updateAccount, deactivateAccount } = useAccounts()
const { setHeaderAction } = useHeaderAction()

const statusFilter = ref<'all' | 'active' | 'inactive'>('active')
const formOpen = ref(false)
const formMode = ref<'create' | 'edit'>('create')
const editingAccount = ref<Account | null>(null)
const isSaving = ref(false)
const formError = ref('')
const confirmOpen = ref(false)
const confirmAccount = ref<Account | null>(null)

const name = ref('')
const type = ref<AccountType>('cash')
const isActive = ref(true)

const { touched, errors, touchField, validateAll } = useInlineValidation(
  { name, type },
  { name: nameSchema(t('accounts.form.nameLabel')), type: accountTypeSchema(t('accounts.form.typeLabel')) }
)

const filteredAccounts = computed(() => {
  if (statusFilter.value === 'active') {
    return accounts.value.filter((item) => item.is_active)
  }
  if (statusFilter.value === 'inactive') {
    return accounts.value.filter((item) => !item.is_active)
  }
  return accounts.value
})

const accountTypeLabel = (value: AccountType) => t(`accounts.types.${value}`)

const statusLabel = (value: boolean) =>
  value ? t('accounts.status.active') : t('accounts.status.inactive')

const resetValidation = () => {
  touched.name = false
  touched.type = false
  errors.name = []
  errors.type = []
}

const resetForm = () => {
  name.value = ''
  type.value = 'cash'
  isActive.value = true
  formError.value = ''
  resetValidation()
}

const openCreate = () => {
  formMode.value = 'create'
  editingAccount.value = null
  resetForm()
  formOpen.value = true
}

setHeaderAction(
  { key: 'accounts:new', labelKey: 'accounts.actions.new', requiresEditor: true },
  openCreate
)

const openEdit = (account: Account) => {
  formMode.value = 'edit'
  editingAccount.value = account
  name.value = account.name
  type.value = account.type
  isActive.value = account.is_active
  formError.value = ''
  resetValidation()
  formOpen.value = true
}

const requestDeactivate = (account: Account) => {
  confirmAccount.value = account
  confirmOpen.value = true
}

const confirmDeactivate = async () => {
  if (!confirmAccount.value) return
  try {
    await deactivateAccount(confirmAccount.value.id)
    push.success({ title: t('accounts.title'), message: t('accounts.messages.deactivated') })
  } catch {
    push.error({ title: t('accounts.title'), message: t('accounts.messages.error') })
  } finally {
    confirmAccount.value = null
  }
}

const handleSubmit = async () => {
  formError.value = ''
  if (validateAll()) {
    return
  }
  isSaving.value = true
  try {
    if (formMode.value === 'create') {
      await createAccount({ name: name.value, type: type.value, is_active: isActive.value })
      push.success({ title: t('accounts.title'), message: t('accounts.messages.created') })
    } else if (editingAccount.value) {
      await updateAccount(editingAccount.value.id, { name: name.value, is_active: isActive.value })
      push.success({ title: t('accounts.title'), message: t('accounts.messages.updated') })
    }
    formOpen.value = false
  } catch (err) {
    if (isApiError(err) && err.code === 'CONFLICT_DUPLICATE_NAME') {
      formError.value = t('accounts.messages.duplicateName')
    } else {
      formError.value = t('accounts.messages.error')
    }
  } finally {
    isSaving.value = false
  }
}

const handleToggleActive = async (account: Account) => {
  if (!canEdit.value) return
  try {
    if (account.is_active) {
      requestDeactivate(account)
      return
    }
    await updateAccount(account.id, { name: account.name, is_active: true })
    push.success({ title: t('accounts.title'), message: t('accounts.messages.activated') })
  } catch {
    push.error({ title: t('accounts.title'), message: t('accounts.messages.error') })
  }
}

watch(
  () => ledgerContext.activeLedgerId.value,
  async (ledgerId) => {
    if (!ledgerId) return
    await fetchAccounts()
  },
  { immediate: true }
)
</script>

<template>
  <div class="grid gap-6">
    <div class="flex flex-wrap items-start justify-between gap-4">
      <TooltipProvider>
        <Tooltip>
          <TooltipTrigger as-child>
            <span>
              <Button :disabled="!canEdit" @click="openCreate">
                <Plus class="h-4 w-4" />
                {{ t('accounts.actions.new') }}
              </Button>
            </span>
          </TooltipTrigger>
          <TooltipContent v-if="!canEdit">
            {{ t('accounts.readOnlyHint') }}
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>
    </div>

    <CrudTableCard
      :title="t('accounts.list.title')"
      :description="t('accounts.list.description')"
      :loading="loading"
      :error="error"
      :empty="filteredAccounts.length === 0"
      :loading-message="t('accounts.loading')"
      :empty-message="t('accounts.empty')"
    >
      <template #toolbar>
        <div class="w-full md:w-56">
          <Label class="text-xs text-muted-foreground">{{ t('accounts.filters.status') }}</Label>
          <Select v-model="statusFilter">
            <SelectTrigger class="mt-2">
              <SelectValue :placeholder="t('accounts.filters.placeholder')" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">{{ t('accounts.filters.all') }}</SelectItem>
              <SelectItem value="active">{{ t('accounts.filters.active') }}</SelectItem>
              <SelectItem value="inactive">{{ t('accounts.filters.inactive') }}</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </template>

      <div class="overflow-hidden rounded-lg border">
        <table class="w-full text-sm">
          <thead class="bg-muted/50 text-xs uppercase text-muted-foreground">
            <tr>
              <th class="px-4 py-3 text-left">{{ t('accounts.table.name') }}</th>
              <th class="px-4 py-3 text-left">{{ t('accounts.table.type') }}</th>
              <th class="px-4 py-3 text-left">{{ t('accounts.table.status') }}</th>
              <th class="px-4 py-3 text-right">{{ t('accounts.table.actions') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="account in filteredAccounts" :key="account.id" class="border-t">
              <td class="px-4 py-3">
                <div class="font-medium text-foreground">{{ account.name }}</div>
                <div class="text-xs text-muted-foreground">{{ account.id }}</div>
              </td>
              <td class="px-4 py-3 text-muted-foreground">
                {{ accountTypeLabel(account.type) }}
              </td>
              <td class="px-4 py-3">
                <span
                  class="inline-flex items-center rounded-full border px-2 py-1 text-xs"
                  :class="account.is_active
                    ? 'border-emerald-500/30 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300'
                    : 'border-muted-foreground/30 bg-muted/50 text-muted-foreground'"
                >
                  {{ statusLabel(account.is_active) }}
                </span>
              </td>
              <td class="px-4 py-3 text-right">
                <DropdownMenu>
                  <DropdownMenuTrigger as-child>
                    <Button variant="ghost" size="icon" :disabled="!canEdit">
                      <MoreHorizontal class="h-4 w-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <DropdownMenuItem @click="openEdit(account)">
                      {{ t('accounts.actions.edit') }}
                    </DropdownMenuItem>
                    <DropdownMenuItem @click="handleToggleActive(account)">
                      {{ account.is_active ? t('accounts.actions.deactivate') : t('accounts.actions.activate') }}
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </CrudTableCard>

    <Sheet v-model:open="formOpen">
      <SheetContent side="right" class="w-full sm:max-w-md px-6 py-6">
        <SheetHeader>
          <SheetTitle>
            {{ formMode === 'create' ? t('accounts.form.createTitle') : t('accounts.form.editTitle') }}
          </SheetTitle>
          <SheetDescription>
            {{ formMode === 'create' ? t('accounts.form.createDescription') : t('accounts.form.editDescription') }}
          </SheetDescription>
        </SheetHeader>

        <div class="mt-6 grid gap-4">
          <div class="grid gap-2">
            <Label for-id="account_name">{{ t('accounts.form.nameLabel') }} <span class="text-destructive">*</span></Label>
            <Input
              id="account_name"
              v-model="name"
              :class="getInputClass({ touched: touched.name, hasError: Boolean(errors.name[0]) })"
              @blur="touchField('name')"
            />
            <p v-if="touched.name && errors.name[0]" class="text-xs text-destructive">
              {{ errors.name[0].message }}
            </p>
          </div>
          <div class="grid gap-2">
            <Label>{{ t('accounts.form.typeLabel') }} <span class="text-destructive">*</span></Label>
            <Select v-model="type" :disabled="formMode === 'edit'">
              <SelectTrigger
                :class="getInputClass({ touched: touched.type, hasError: Boolean(errors.type[0]) })"
                @blur="touchField('type')"
              >
                <SelectValue :placeholder="t('accounts.form.typePlaceholder')" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="cash">{{ t('accounts.types.cash') }}</SelectItem>
                <SelectItem value="investment">{{ t('accounts.types.investment') }}</SelectItem>
                <SelectItem value="credit_card">{{ t('accounts.types.credit_card') }}</SelectItem>
              </SelectContent>
            </Select>
            <p v-if="touched.type && errors.type[0]" class="text-xs text-destructive">
              {{ errors.type[0].message }}
            </p>
          </div>
          <p v-if="formError" class="text-sm text-destructive">
            {{ formError }}
          </p>
        </div>

        <SheetFooter class="mt-6 flex-row justify-end gap-2">
          <Button variant="outline" type="button" @click="formOpen = false">
            {{ t('accounts.form.cancel') }}
          </Button>
          <Button type="button" :disabled="isSaving" @click="handleSubmit">
            {{ isSaving ? t('accounts.form.saving') : t('accounts.form.submit') }}
          </Button>
        </SheetFooter>
      </SheetContent>
    </Sheet>

    <ConfirmDialog
      v-model:open="confirmOpen"
      :title="t('accounts.confirm.deactivateTitle')"
      :description="confirmAccount ? t('accounts.confirm.deactivateMessage', { name: confirmAccount.name }) : ''"
      :confirm-label="t('accounts.confirm.confirmAction')"
      :cancel-label="t('accounts.confirm.cancelAction')"
      @confirm="confirmDeactivate"
      @cancel="confirmOpen = false"
    />
  </div>
</template>
