<script setup lang="ts">
import { computed, ref } from 'vue'
import { Button } from '@shared/components/ui/button'
import { Card, CardContent } from '@shared/components/ui/card'
import { Input } from '@shared/components/ui/input'
import { Label } from '@shared/components/ui/label'
import { Separator } from '@shared/components/ui/separator'
import googleIcon from '@shared/icons/google.svg'
import githubIcon from '@shared/icons/github.svg'
import { useAuth } from '#layers/auth/composables/useAuth'
import { emailSchema, nameSchema, passwordSchema, useInlineValidation, getInputClass } from '@shared/validators'

const displayName = ref('')
const email = ref('')
const password = ref('')
const passwordFocused = ref(false)
const error = ref('')
const isLoading = ref(false)

const { signup, startOAuth, extractErrorMessage, extractErrorDetails, extractErrorCode } = useAuth()
const router = useRouter()

const { touched, errors, touchField, validateAll } = useInlineValidation(
  { displayName, email, password },
  { displayName: nameSchema(), email: emailSchema(), password: passwordSchema() }
)

const oauthLoading = ref<'google' | 'github' | null>(null)

const inputClass = (field: keyof typeof errors) =>
  getInputClass({ touched: touched[field], hasError: Boolean(errors[field][0]) })

const passwordHints = computed(() => ({
  minLength: password.value.length >= 8,
  uppercase: /[A-Z]/.test(password.value),
  number: /[0-9]/.test(password.value)
}))

const onSubmit = async () => {
  error.value = ''
  if (validateAll()) {
    return
  }
  isLoading.value = true
  try {
    await signup(email.value, password.value, displayName.value)
    await router.push('/')
  } catch (err) {
    const code = extractErrorCode(err)
    if (code === 'VALIDATION_ERROR') {
      const details = extractErrorDetails(err)
      if (details.email === 'invalid') {
        error.value = 'E-mail inválido.'
        return
      }
      if (details.password === 'min_length') {
        error.value = 'A senha deve ter pelo menos 8 caracteres.'
        return
      }
    }
    error.value = extractErrorMessage(err)
  } finally {
    isLoading.value = false
  }
}

const handleOAuth = (provider: 'google' | 'github') => {
  oauthLoading.value = provider
  startOAuth(provider)
}
</script>

<template>
  <div class="flex flex-col gap-6">
    <Card class="p-0">
      <CardContent class="grid p-0 md:grid-cols-2">
        <form class="p-6 md:p-8" @submit.prevent="onSubmit">
          <div class="flex flex-col gap-6">
            <div class="flex flex-col items-center gap-2 text-center">
              <h1 class="text-2xl font-bold">Criar conta</h1>
              <p class="text-balance text-sm text-muted-foreground">
                Crie seu ledger HausHaltsMeister em minutos
              </p>
            </div>
            <div class="grid gap-2">
              <Label for-id="display_name">Nome <span class="text-destructive">*</span></Label>
              <Input
                id="display_name"
                v-model="displayName"
                type="text"
                autocomplete="name"
                :class="inputClass('displayName')"
                @blur="touchField('displayName')"
              />
              <p v-if="touched.displayName && errors.displayName[0]" class="text-xs text-destructive">
                {{ errors.displayName[0].message }}
              </p>
            </div>
            <div class="grid gap-2">
              <Label for-id="email">E-mail <span class="text-destructive">*</span></Label>
              <Input
                id="email"
                v-model="email"
                type="email"
                autocomplete="email"
                placeholder="email@exemplo.com"
                :class="inputClass('email')"
                @blur="touchField('email')"
              />
              <p v-if="touched.email && errors.email[0]" class="text-xs text-destructive">
                {{ errors.email[0].message }}
              </p>
            </div>
            <div class="grid gap-2 relative">
              <Label for-id="password">Senha <span class="text-destructive">*</span></Label>
              <Input
                id="password"
                v-model="password"
                type="password"
                autocomplete="new-password"
                :class="inputClass('password')"
                @focus="passwordFocused = true"
                @blur="passwordFocused = false; touchField('password')"
              />
              <p class="text-xs text-muted-foreground">Mínimo de 8 caracteres.</p>
              <p v-if="touched.password && errors.password[0]" class="text-xs text-destructive">
                {{ errors.password[0].message }}
              </p>
              <div
                v-if="passwordFocused"
                class="absolute top-1/2 right-full z-10 mr-4 w-72 -translate-y-1/2 rounded-lg border bg-card p-4 text-sm shadow-lg"
                :class="errors.password[0] ? 'border-destructive/40' : 'border-primary/30'"
              >
                <div class="font-medium text-foreground">Requisitos da senha</div>
                <div class="mt-1 text-muted-foreground">Recomendamos:</div>
                <div class="mt-3 space-y-2">
                  <div class="flex items-center gap-2">
                    <span
                      class="inline-flex h-2.5 w-2.5 rounded-full"
                      :class="passwordHints.minLength ? 'bg-primary' : 'bg-muted-foreground/40'"
                    />
                    <span>Mínimo de 8 caracteres</span>
                  </div>
                  <div class="flex items-center gap-2">
                    <span
                      class="inline-flex h-2.5 w-2.5 rounded-full"
                      :class="passwordHints.uppercase ? 'bg-primary' : 'bg-muted-foreground/40'"
                    />
                    <span>Ao menos 1 letra maiúscula</span>
                  </div>
                  <div class="flex items-center gap-2">
                    <span
                      class="inline-flex h-2.5 w-2.5 rounded-full"
                      :class="passwordHints.number ? 'bg-primary' : 'bg-muted-foreground/40'"
                    />
                    <span>Ao menos 1 número</span>
                  </div>
                </div>
              </div>
            </div>
            <Button type="submit" class="w-full" :disabled="isLoading">
              {{ isLoading ? 'Criando...' : 'Criar conta' }}
            </Button>
            <p v-if="error" class="text-center text-sm text-destructive">
              {{ error }}
            </p>
            <Separator>Ou continue com</Separator>
            <div class="grid grid-cols-2 gap-4">
              <Button variant="outline" type="button" :disabled="oauthLoading !== null" @click="handleOAuth('google')">
                <span
                  v-if="oauthLoading === 'google'"
                  class="h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent"
                />
                <img v-else :src="googleIcon" alt="" class="h-4 w-4" />
                <span>Google</span>
              </Button>
              <Button variant="outline" type="button" :disabled="oauthLoading !== null" @click="handleOAuth('github')">
                <span
                  v-if="oauthLoading === 'github'"
                  class="h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent"
                />
                <img v-else :src="githubIcon" alt="" class="h-4 w-4" />
                <span>GitHub</span>
              </Button>
            </div>
            <div class="text-center text-sm text-muted-foreground">
              Já tem conta?
              <NuxtLink to="/auth/login" class="underline underline-offset-4">Entrar</NuxtLink>
            </div>
          </div>
        </form>
        <div class="bg-muted relative hidden overflow-hidden md:block">
          <img
            src="/placeholder.svg"
            alt="Image"
            class="absolute inset-0 h-full w-full object-cover dark:brightness-[0.2] dark:grayscale"
          />
        </div>
      </CardContent>
    </Card>
    <p class="px-6 text-center text-xs text-muted-foreground">
      Ao continuar, você concorda com nossos
      <a href="#" class="underline underline-offset-4">Termos de Uso</a>
      e
      <a href="#" class="underline underline-offset-4">Política de Privacidade</a>.
    </p>
  </div>
</template>
