<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import EntryTable from '../../components/EntryTable.vue'
import EntryFormSheet from '../../components/EntryFormSheet.vue'
import EntryDeleteDialog from '../../components/EntryDeleteDialog.vue'
import type { AddEntryRequest, PaymentMethod, PicuinhaEntry, Person } from '../../types/picuinha'
import { usePicuinhasService } from '../../services/picuinhas'
import { getApiErrorMessage } from '~/layers/shared/utils/api'
import { Button } from '~/layers/shared/components/ui/button'

definePageMeta({
  layout: 'default',
})

const { listPersons, addEntry, listEntries, updateEntry, deleteEntry, listPaymentMethods } = usePicuinhasService()

const entries = ref<PicuinhaEntry[]>([])
const loading = ref(true)
const loadError = ref<string | null>(null)

const persons = ref<Person[]>([])
const personsLoading = ref(false)
const personsError = ref<string | null>(null)

const paymentMethods = ref<PaymentMethod[]>([])
const paymentMethodsLoading = ref(false)
const paymentMethodsError = ref<string | null>(null)

const formOpen = ref(false)
const formMode = ref<'create' | 'edit'>('create')
const selectedEntry = ref<PicuinhaEntry | null>(null)
const formError = ref<string | null>(null)
const submitting = ref(false)

const deleteOpen = ref(false)
const deleteTarget = ref<PicuinhaEntry | null>(null)
const deleting = ref(false)

const feedback = ref<{ type: 'success' | 'error'; message: string } | null>(null)

const entriesSupported = true
const entryActionsDisabled = false

const pageSubtitle = computed(() => 'Registre empréstimos e recebimentos vinculados às pessoas.')

const entryRows = computed(() => {
  const personMap = new Map(persons.value.map((person) => [person.id, person.name]))
  const paymentMap = new Map(paymentMethods.value.map((method) => [method.id, method.name]))
  return entries.value.map((entry) => ({
    ...entry,
    person_name: personMap.get(entry.person_id),
    payment_method_name: entry.payment_method_id ? paymentMap.get(entry.payment_method_id) : undefined,
  }))
})

async function fetchPersons() {
  personsLoading.value = true
  personsError.value = null
  try {
    persons.value = await listPersons()
  } catch (error) {
    personsError.value = getApiErrorMessage(error)
  } finally {
    personsLoading.value = false
  }
}

async function fetchPaymentMethods() {
  paymentMethodsLoading.value = true
  paymentMethodsError.value = null
  try {
    paymentMethods.value = await listPaymentMethods()
  } catch (error) {
    paymentMethodsError.value = getApiErrorMessage(error)
  } finally {
    paymentMethodsLoading.value = false
  }
}

async function fetchEntries() {
  loading.value = true
  loadError.value = null
  try {
    entries.value = await listEntries()
  } catch (error) {
    loadError.value = getApiErrorMessage(error)
  } finally {
    loading.value = false
  }
}

async function refreshData() {
  await Promise.all([fetchPersons(), fetchEntries(), fetchPaymentMethods()])
}

function openCreate() {
  formMode.value = 'create'
  selectedEntry.value = null
  formError.value = null
  formOpen.value = true
}

function openEdit(entry: PicuinhaEntry) {
  formMode.value = 'edit'
  selectedEntry.value = entry
  formError.value = null
  formOpen.value = true
}

async function handleSubmit(payload: AddEntryRequest) {
  submitting.value = true
  formError.value = null
  try {
    if (formMode.value === 'edit' && selectedEntry.value) {
      const updated = await updateEntry(selectedEntry.value.id, payload)
      entries.value = entries.value.map((item) => (item.id === updated.id ? updated : item))
      feedback.value = { type: 'success', message: 'Lançamento atualizado com sucesso.' }
    } else {
      await addEntry(payload)
      feedback.value = { type: 'success', message: 'Lançamento criado com sucesso.' }
      if (entriesSupported) {
        await fetchEntries()
      }
    }
    formOpen.value = false
  } catch (error) {
    formError.value = getApiErrorMessage(error)
  } finally {
    submitting.value = false
  }
}

function requestDelete(entry: PicuinhaEntry) {
  deleteTarget.value = entry
  deleteOpen.value = true
}

async function confirmDelete() {
  if (!deleteTarget.value) return
  deleting.value = true
  try {
    await deleteEntry(deleteTarget.value.id)
    entries.value = entries.value.filter((item) => item.id !== deleteTarget.value?.id)
    feedback.value = { type: 'success', message: 'Lançamento excluído com sucesso.' }
    deleteOpen.value = false
  } catch (error) {
    feedback.value = { type: 'error', message: getApiErrorMessage(error) }
  } finally {
    deleting.value = false
  }
}

watch(formOpen, (open) => {
  if (!open) {
    formError.value = null
  }
})

onMounted(async () => {
  await refreshData()
})
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
      <div class="space-y-1">
        <h1 class="text-2xl font-semibold tracking-tight">Picuinhas · Lançamentos</h1>
        <p class="text-sm text-muted-foreground">{{ pageSubtitle }}</p>
      </div>
      <Button variant="outline" :disabled="personsLoading" @click="refreshData">Atualizar lista</Button>
    </div>

    <div
      v-if="feedback"
      class="rounded-md border px-4 py-3 text-sm"
      :class="feedback.type === 'error' ? 'border-destructive/30 bg-destructive/10 text-destructive' : 'border-primary/30 bg-primary/10 text-primary'"
    >
      <div class="flex items-center justify-between gap-4">
        <span>{{ feedback.message }}</span>
        <Button variant="ghost" size="sm" @click="feedback = null">Fechar</Button>
      </div>
    </div>

    <div v-if="personsError" class="rounded-md border border-destructive/30 bg-destructive/10 p-3 text-sm text-destructive">
      {{ personsError }}
    </div>

    <div v-if="paymentMethodsError" class="rounded-md border border-destructive/30 bg-destructive/10 p-3 text-sm text-destructive">
      {{ paymentMethodsError }}
    </div>

    <EntryTable
      :entries="entryRows"
      :loading="loading || personsLoading || paymentMethodsLoading"
      :error="loadError"
      :actions-disabled="entryActionsDisabled"
      @create="openCreate"
      @edit="openEdit"
      @remove="requestDelete"
      @retry="fetchEntries"
    />

    <EntryFormSheet
      v-model:open="formOpen"
      :mode="formMode"
      :entry="selectedEntry"
      :persons="persons"
      :payment-methods="paymentMethods"
      :submitting="submitting"
      :error-message="formError"
      @submit="handleSubmit"
    />

    <EntryDeleteDialog
      v-model:open="deleteOpen"
      :entry="deleteTarget"
      :submitting="deleting"
      @confirm="confirmDelete"
    />
  </div>
</template>
