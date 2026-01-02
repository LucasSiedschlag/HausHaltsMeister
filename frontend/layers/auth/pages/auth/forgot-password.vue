<script setup lang="ts">
import { ref } from 'vue'
import { Button } from '@shared/components/ui/button'
import { Input } from '@shared/components/ui/input'
import { Label } from '@shared/components/ui/label'
import { useAuth } from '#layers/auth/composables/useAuth'
import { emailSchema, useInlineValidation, getInputClass } from '@shared/validators'
definePageMeta({ layout: false })

const email = ref('')
const error = ref('')
const success = ref('')
const isLoading = ref(false)
const { touched, errors, touchField, validateAll } = useInlineValidation({ email }, { email: emailSchema() })

const { forgotPassword, extractErrorMessage, extractErrorDetails, extractErrorCode } = useAuth()

const onSubmit = async () => {
  error.value = ''
  success.value = ''
  if (validateAll()) {
    return
  }
  isLoading.value = true
  try {
    await forgotPassword(email.value)
    success.value = 'Se o e-mail existir, enviaremos instruções.'
  } catch (err) {
    const code = extractErrorCode(err)
    if (code === 'VALIDATION_ERROR') {
      const details = extractErrorDetails(err)
      if (details.email === 'invalid') {
        error.value = 'E-mail inválido.'
        return
      }
    }
    error.value = extractErrorMessage(err)
  } finally {
    isLoading.value = false
  }
}
</script>

<template>
  <div class="grid gap-6">
    <div class="grid gap-2 text-center">
      <h1 class="text-3xl font-semibold tracking-tight">Recuperar senha</h1>
      <p class="text-sm text-muted-foreground">Enviaremos um link para seu e-mail</p>
    </div>
    <div class="grid gap-4">
      <form class="grid gap-4" @submit.prevent="onSubmit">
        <div class="grid gap-2">
          <Label for-id="email">E-mail <span class="text-destructive">*</span></Label>
          <Input
            id="email"
            v-model="email"
            type="email"
            autocomplete="email"
            placeholder="email@exemplo.com"
            :class="getInputClass({ touched: touched.email, hasError: Boolean(errors.email[0]) })"
            @blur="touchField('email')"
          />
          <p v-if="touched.email && errors.email[0]" class="text-xs text-destructive">
            {{ errors.email[0].message }}
          </p>
        </div>
        <p v-if="error" class="rounded-md border border-destructive/40 bg-destructive/10 px-3 py-2 text-sm text-destructive">
          {{ error }}
        </p>
        <p v-if="success" class="rounded-md border border-primary/30 bg-primary/5 px-3 py-2 text-sm text-primary">
          {{ success }}
        </p>
        <Button class="w-full" type="submit" :disabled="isLoading">
          {{ isLoading ? 'Enviando...' : 'Enviar' }}
        </Button>
      </form>
    </div>
    <div class="text-center text-sm text-muted-foreground">
      Lembrou da senha?
      <NuxtLink to="/auth/login" class="underline underline-offset-4">Voltar ao login</NuxtLink>
    </div>
  </div>
</template>
