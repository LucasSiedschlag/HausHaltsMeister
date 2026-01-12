<script setup lang="ts">
import { computed, ref, watch, onMounted } from 'vue'
import { Button } from '@shared/components/ui/button'
import { Input } from '@shared/components/ui/input'
import { Label } from '@shared/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@shared/components/ui/select'
import { Sheet, SheetContent, SheetDescription, SheetFooter, SheetHeader, SheetTitle } from '@shared/components/ui/sheet'
import { push } from 'notivue'
import { useLedgerContext } from '@shared/composables/useLedgerContext'
import { useHeaderAction } from '@shared/composables/useHeaderAction'
import { useAccounts } from '#layers/accounts/composables/useAccounts'
import { useCreditCards, type CreditCard } from '#layers/creditcard/composables/useCreditCards'
import { useCardNetworks } from '#layers/creditcard/composables/useCardNetworks'
import { integerSchema, nameSchema, getInputClass, useInlineValidation } from '@shared/validators'
import { isApiError } from '@shared/utils/api-error'
import ConfirmDialog from '@shared/components/ConfirmDialog.vue'
import CreditCardTile from '#layers/creditcard/components/CreditCardTile.vue'

const { t, te } = useI18n()
const ledgerContext = useLedgerContext()
const canEdit = computed(() => ledgerContext.hasRole('editor'))
const { setHeaderAction } = useHeaderAction()

const { accounts, fetchAccounts } = useAccounts()
const { cards, loading, error, fetchCards, createCard, updateCard, deleteCard } = useCreditCards()
const { networks, fetchNetworks } = useCardNetworks()

const formOpen = ref(false)
const formMode = ref<'create' | 'edit'>('create')
const editingCard = ref<CreditCard | null>(null)
const isSaving = ref(false)
const formError = ref('')

const accountId = ref('')
const label = ref('')
const holderName = ref('')
const networkCode = ref('')
const last4 = ref('')
const closingDay = ref('')
const dueDay = ref('')

const { touched, errors, touchField, validateAll } = useInlineValidation(
  { accountId, networkCode, closingDay, dueDay },
  {
    accountId: nameSchema(t('creditCards.form.accountLabel')),
    networkCode: nameSchema(t('creditCards.form.networkLabel')),
    closingDay: integerSchema(t('creditCards.form.closingDayLabel'), { min: 1, max: 28 }),
    dueDay: integerSchema(t('creditCards.form.dueDayLabel'), { min: 1, max: 28 })
  }
)

const confirmOpen = ref(false)
const confirmCard = ref<CreditCard | null>(null)

const eligibleAccounts = computed(() => accounts.value)
const accountNameMap = computed(() =>
  new Map(accounts.value.map((account) => [account.id, account.name]))
)

const availableAccounts = computed(() => {
  if (formMode.value === 'edit') return eligibleAccounts.value
  return eligibleAccounts.value.filter((account) => account.is_active)
})

const sortedCards = computed(() => {
  return [...cards.value].sort((a, b) => {
    const nameA = (a.label || accountNameMap.value.get(a.parent_account_id) || '').toLowerCase()
    const nameB = (b.label || accountNameMap.value.get(b.parent_account_id) || '').toLowerCase()
    return nameA.localeCompare(nameB)
  })
})

const networkNameMap = computed(() =>
  new Map(networks.value.map((network) => [network.code, network.display_name]))
)

const resetValidation = () => {
  touched.accountId = false
  touched.networkCode = false
  touched.closingDay = false
  touched.dueDay = false
  errors.accountId = []
  errors.networkCode = []
  errors.closingDay = []
  errors.dueDay = []
}

const resetForm = () => {
  accountId.value = ''
  label.value = ''
  holderName.value = ''
  networkCode.value = ''
  last4.value = ''
  closingDay.value = ''
  dueDay.value = ''
  formError.value = ''
  resetValidation()
}

const openCreate = () => {
  formMode.value = 'create'
  editingCard.value = null
  resetForm()
  accountId.value = availableAccounts.value[0]?.id ?? ''
  networkCode.value = networks.value[0]?.code ?? ''
  formOpen.value = true
}

const openEdit = (card: CreditCard) => {
  formMode.value = 'edit'
  editingCard.value = card
  accountId.value = card.parent_account_id
  label.value = card.label || ''
  holderName.value = card.holder_name || ''
  networkCode.value = card.brand
  last4.value = card.last4 || ''
  closingDay.value = String(card.closing_day)
  dueDay.value = String(card.due_day)
  formError.value = ''
  resetValidation()
  formOpen.value = true
}

const handleSubmit = async () => {
  formError.value = ''
  if (validateAll()) return
  if (last4.value && !/^\d{4}$/.test(last4.value.trim())) {
    formError.value = t('creditCards.messages.last4Invalid')
    return
  }
  isSaving.value = true
  try {
    const payload = {
      label: label.value.trim() || null,
      holder_name: holderName.value.trim() || null,
      brand: networkCode.value,
      last4: last4.value.trim() || null,
      closing_day: Number(closingDay.value),
      due_day: Number(dueDay.value)
    }
    if (formMode.value === 'create') {
      await createCard(accountId.value, payload)
      push.success({ title: t('creditCards.title'), message: t('creditCards.messages.created') })
    } else if (editingCard.value) {
      await updateCard(editingCard.value.id, payload)
      push.success({ title: t('creditCards.title'), message: t('creditCards.messages.updated') })
    }
    formOpen.value = false
  } catch (err) {
    if (isApiError(err) && te(`common.errors.${err.code}`)) {
      formError.value = t(`common.errors.${err.code}`)
    } else {
      formError.value = t('creditCards.messages.error')
    }
  } finally {
    isSaving.value = false
  }
}

const requestDelete = (card: CreditCard) => {
  confirmCard.value = card
  confirmOpen.value = true
}

const confirmDelete = async () => {
  if (!confirmCard.value) return
  try {
    await deleteCard(confirmCard.value.id)
    push.success({ title: t('creditCards.title'), message: t('creditCards.messages.deleted') })
  } catch {
    push.error({ title: t('creditCards.title'), message: t('creditCards.messages.error') })
  } finally {
    confirmCard.value = null
  }
}

setHeaderAction(
  { key: 'credit-cards:new', labelKey: 'creditCards.header.newCard', requiresEditor: true },
  openCreate
)

watch(
  () => ledgerContext.activeLedgerId.value,
  async (ledgerId) => {
    if (!ledgerId) return
    await fetchAccounts()
    await Promise.all(accounts.value.map((account) => fetchCards(account.id)))
  },
  { immediate: true }
)

onMounted(async () => {
  await fetchNetworks()
})
</script>

<template>
  <div class="grid gap-6">
    <div v-if="loading" class="text-sm text-muted-foreground">
      {{ t('creditCards.loading') }}
    </div>
    <div v-else-if="error" class="text-sm text-destructive">
      {{ error }}
    </div>
    <div v-else-if="sortedCards.length === 0" class="space-y-2 text-sm text-muted-foreground">
      <div>{{ t('creditCards.empty') }}</div>
      <div v-if="eligibleAccounts.length === 0" class="text-xs text-muted-foreground">
        {{ t('creditCards.list.emptyAccounts') }}
      </div>
    </div>
    <div v-else class="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
      <CreditCardTile
        v-for="card in sortedCards"
        :key="card.id"
        :card="card"
        :account-name="accountNameMap.get(card.parent_account_id) || card.parent_account_id"
        :network-name="networkNameMap.get(card.brand) || card.brand"
        :to="`/credit-cards/${card.id}`"
        :can-edit="canEdit"
        @edit="openEdit"
        @delete="requestDelete"
      />
    </div>

    <Sheet v-model:open="formOpen">
      <SheetContent side="right" class="w-full sm:max-w-lg px-6 py-6">
        <SheetHeader>
          <SheetTitle>
            {{ formMode === 'create' ? t('creditCards.form.createTitle') : t('creditCards.form.editTitle') }}
          </SheetTitle>
          <SheetDescription>
            {{ formMode === 'create' ? t('creditCards.form.createDescription') : t('creditCards.form.editDescription') }}
          </SheetDescription>
        </SheetHeader>

        <div class="mt-6 space-y-4">
          <div class="space-y-2">
            <Label>{{ t('creditCards.form.accountLabel') }}</Label>
            <Select v-model="accountId" :disabled="formMode === 'edit'">
            <SelectTrigger :class="getInputClass({ touched: touched.accountId, hasError: Boolean(errors.accountId?.[0]) })">
                <SelectValue :placeholder="t('creditCards.form.accountPlaceholder')" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem
                  v-for="account in availableAccounts"
                  :key="account.id"
                  :value="account.id"
                >
                  {{ account.name }}
                </SelectItem>
              </SelectContent>
            </Select>
            <p v-if="touched.accountId && errors.accountId.length" class="text-xs text-destructive">
              {{ errors.accountId?.[0]?.message }}
            </p>
          </div>

          <div class="space-y-2">
            <Label>{{ t('creditCards.form.networkLabel') }}</Label>
            <Select v-model="networkCode">
            <SelectTrigger :class="getInputClass({ touched: touched.networkCode, hasError: Boolean(errors.networkCode?.[0]) })">
                <SelectValue :placeholder="t('creditCards.form.networkPlaceholder')" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem
                  v-for="network in networks"
                  :key="network.code"
                  :value="network.code"
                >
                  {{ network.display_name }}
                </SelectItem>
              </SelectContent>
            </Select>
            <p v-if="touched.networkCode && errors.networkCode.length" class="text-xs text-destructive">
              {{ errors.networkCode?.[0]?.message }}
            </p>
          </div>

          <div class="grid gap-4 sm:grid-cols-2">
            <div class="space-y-2">
              <Label>{{ t('creditCards.form.labelLabel') }}</Label>
              <Input v-model="label" :placeholder="t('creditCards.form.labelPlaceholder')" />
            </div>
            <div class="space-y-2">
              <Label>{{ t('creditCards.form.holderLabel') }}</Label>
              <Input v-model="holderName" :placeholder="t('creditCards.form.holderPlaceholder')" />
            </div>
          </div>

          <div class="space-y-2">
            <Label>{{ t('creditCards.form.last4Label') }}</Label>
            <Input v-model="last4" maxlength="4" :placeholder="t('creditCards.form.last4Placeholder')" />
          </div>

          <div class="grid gap-4 sm:grid-cols-2">
            <div class="space-y-2">
              <Label>{{ t('creditCards.form.closingDayLabel') }}</Label>
              <Input
                v-model="closingDay"
                inputmode="numeric"
                :placeholder="t('creditCards.form.closingDayPlaceholder')"
                :class="getInputClass({ touched: touched.closingDay, hasError: Boolean(errors.closingDay?.[0]) })"
                @blur="touchField('closingDay')"
              />
              <p v-if="touched.closingDay && errors.closingDay.length" class="text-xs text-destructive">
                {{ errors.closingDay?.[0]?.message }}
              </p>
            </div>
            <div class="space-y-2">
              <Label>{{ t('creditCards.form.dueDayLabel') }}</Label>
              <Input
                v-model="dueDay"
                inputmode="numeric"
                :placeholder="t('creditCards.form.dueDayPlaceholder')"
                :class="getInputClass({ touched: touched.dueDay, hasError: Boolean(errors.dueDay?.[0]) })"
                @blur="touchField('dueDay')"
              />
              <p v-if="touched.dueDay && errors.dueDay.length" class="text-xs text-destructive">
                {{ errors.dueDay?.[0]?.message }}
              </p>
            </div>
          </div>

          <p v-if="formError" class="text-sm text-destructive">
            {{ formError }}
          </p>
        </div>

        <SheetFooter class="mt-6">
          <Button variant="outline" @click="formOpen = false">
            {{ t('creditCards.form.cancel') }}
          </Button>
          <Button :disabled="isSaving" @click="handleSubmit">
            {{ isSaving ? t('creditCards.form.saving') : t('creditCards.form.submit') }}
          </Button>
        </SheetFooter>
      </SheetContent>
    </Sheet>

    <ConfirmDialog
      v-model:open="confirmOpen"
      :title="t('creditCards.confirm.deleteTitle')"
      :description="t('creditCards.confirm.deleteMessage', { name: confirmCard?.label || accountNameMap.get(confirmCard?.parent_account_id || '') || '' })"
      :confirm-label="t('creditCards.confirm.confirmAction')"
      :cancel-label="t('creditCards.confirm.cancelAction')"
      @confirm="confirmDelete"
      @cancel="confirmCard = null"
    />

  </div>
</template>
