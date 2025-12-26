<script setup lang="ts">
import { computed, onActivated, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import PersonTable from '../../components/PersonTable.vue'
import PersonFormSheet from '../../components/PersonFormSheet.vue'
import PersonDeleteDialog from '../../components/PersonDeleteDialog.vue'
import type { CreatePersonRequest, Person, UpdatePersonRequest } from '../../types/picuinha'
import { usePicuinhasService } from '../../services/picuinhas'
import { getApiErrorMessage } from '~/layers/shared/utils/api'
import { Button } from '~/layers/shared/components/ui/button'

definePageMeta({
  layout: 'default',
})

const { listPersons, createPerson, updatePerson, deletePerson } = usePicuinhasService()
const router = useRouter()

const persons = ref<Person[]>([])
const loading = ref(true)
const loadError = ref<string | null>(null)

const formOpen = ref(false)
const formMode = ref<'create' | 'edit'>('create')
const selectedPerson = ref<Person | null>(null)
const formError = ref<string | null>(null)
const submitting = ref(false)

const deleteOpen = ref(false)
const deleteTarget = ref<Person | null>(null)
const deleting = ref(false)

const feedback = ref<{ type: 'success' | 'error'; message: string } | null>(null)

const pageSubtitle = computed(() => 'Cadastro de pessoas para controle de picuinhas.')

async function fetchPersons() {
  loading.value = true
  loadError.value = null
  try {
    persons.value = await listPersons()
  } catch (error) {
    loadError.value = getApiErrorMessage(error)
  } finally {
    loading.value = false
  }
}

function openCreate() {
  formMode.value = 'create'
  selectedPerson.value = null
  formError.value = null
  formOpen.value = true
}

function openEdit(person: Person) {
  formMode.value = 'edit'
  selectedPerson.value = person
  formError.value = null
  formOpen.value = true
}

async function handleSubmit(payload: UpdatePersonRequest) {
  submitting.value = true
  formError.value = null
  try {
    if (formMode.value === 'edit' && selectedPerson.value) {
      const updated = await updatePerson(selectedPerson.value.id, payload)
      persons.value = persons.value.map((item) => (item.id === updated.id ? updated : item))
      feedback.value = { type: 'success', message: 'Pessoa atualizada com sucesso.' }
    } else {
      const createPayload: CreatePersonRequest = {
        name: payload.name,
        notes: payload.notes,
      }
      const created = await createPerson(createPayload)
      persons.value = [created, ...persons.value]
      feedback.value = { type: 'success', message: 'Pessoa criada com sucesso.' }
    }
    formOpen.value = false
  } catch (error) {
    formError.value = getApiErrorMessage(error)
  } finally {
    submitting.value = false
  }
}

function requestDelete(person: Person) {
  deleteTarget.value = person
  deleteOpen.value = true
}

function viewPersonCases(person: Person) {
  router.push({ path: '/picuinhas/lancamentos', query: { person_id: String(person.id) } })
}

async function confirmDelete() {
  if (!deleteTarget.value) return
  deleting.value = true
  try {
    await deletePerson(deleteTarget.value.id)
    persons.value = persons.value.filter((item) => item.id !== deleteTarget.value?.id)
    feedback.value = { type: 'success', message: 'Pessoa excluída com sucesso.' }
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

onMounted(fetchPersons)
onActivated(fetchPersons)
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
      <div class="space-y-1">
        <h1 class="text-2xl font-semibold tracking-tight">Picuinhas · Pessoas</h1>
        <p class="text-sm text-muted-foreground">{{ pageSubtitle }}</p>
      </div>
      <Button variant="outline" @click="fetchPersons">Atualizar lista</Button>
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

    <PersonTable
      :persons="persons"
      :loading="loading"
      :error="loadError"
      @create="openCreate"
      @edit="openEdit"
      @remove="requestDelete"
      @view="viewPersonCases"
      @retry="fetchPersons"
    />

    <PersonFormSheet
      v-model:open="formOpen"
      :mode="formMode"
      :person="selectedPerson"
      :submitting="submitting"
      :error-message="formError"
      @submit="handleSubmit"
    />

    <PersonDeleteDialog
      v-model:open="deleteOpen"
      :person="deleteTarget"
      :submitting="deleting"
      @confirm="confirmDelete"
    />
  </div>
</template>
