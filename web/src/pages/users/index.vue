<template>
  <div class="users-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ $t('users.title') }}</span>
          <el-button
            type="primary"
            size="small"
            @click="showCreate = true"
          >
            {{ $t('users.createBtn') }}
          </el-button>
        </div>
      </template>
      <el-table
        v-loading="loading"
        :data="users"
        stripe
        style="width: 100%"
      >
        <el-table-column
          prop="name"
          :label="$t('users.name')"
        />
        <el-table-column
          prop="email"
          :label="$t('users.email')"
        />
        <el-table-column
          prop="role"
          :label="$t('users.role')"
        >
          <template #default="{ row }">
            <el-tag
              v-if="row.role === 'super_admin'"
              type="danger"
              size="small"
            >
              {{ $t('users.roleSuperAdmin') }}
            </el-tag>
            <el-tag
              v-else-if="row.role === 'admin'"
              type="warning"
              size="small"
            >
              {{ $t('users.roleAdmin') }}
            </el-tag>
            <el-tag
              v-else
              type="primary"
              size="small"
            >
              {{ $t('users.roleAgent') }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          prop="status"
          :label="$t('users.status')"
        >
          <template #default="{ row }">
            <el-tag
              :type="row.status === 'active' ? 'success' : 'info'"
              size="small"
            >
              {{ row.status === 'active' ? $t('users.statusActive') : $t('users.statusDisabled') }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          prop="must_change_password"
          :label="$t('users.passwordStatus')"
          width="120"
        >
          <template #default="{ row }">
            <el-tag
              :type="row.must_change_password ? 'warning' : 'success'"
              size="small"
            >
              {{ row.must_change_password ? $t('users.pwdNeedChange') : $t('users.pwdSet') }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          prop="created_at"
          :label="$t('users.createdAt')"
          width="170"
        >
          <template #default="{ row }">
            {{ new Date(row.created_at).toLocaleString(locale === 'zh-CN' ? 'zh-CN' : 'en-US') }}
          </template>
        </el-table-column>
        <el-table-column
          :label="$t('users.actions')"
          width="240"
        >
          <template #default="{ row }">
            <el-button
              size="small"
              @click="editUser(row)"
            >
              {{ $t('users.editBtn') }}
            </el-button>
            <el-button
              v-if="userStore.isSuperAdmin()"
              size="small"
              type="warning"
              @click="handleReset(row)"
            >
              {{ $t('users.resetPwd') }}
            </el-button>
            <el-button
              size="small"
              type="danger"
              @click="handleDelete(row)"
            >
              {{ $t('users.deleteBtn') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog
      v-model="showCreate"
      :title="$t('users.createTitle')"
      width="460px"
    >
      <el-form
        ref="createFormRef"
        :model="createForm"
        :rules="createRules"
        label-width="80px"
      >
        <el-form-item
          :label="$t('users.createEmail')"
          prop="email"
        >
          <el-input v-model="createForm.email" />
        </el-form-item>
        <el-form-item
          :label="$t('users.createName')"
          prop="name"
        >
          <el-input v-model="createForm.name" />
        </el-form-item>
        <el-form-item
          :label="$t('users.createPassword')"
          prop="password"
        >
          <el-input
            v-model="createForm.password"
            type="password"
            show-password
          />
        </el-form-item>
        <el-form-item
          :label="$t('users.createRole')"
          prop="role"
        >
          <el-select
            v-model="createForm.role"
            style="width: 100%"
          >
            <el-option
              :label="$t('users.roleAgent')"
              value="agent"
            />
            <el-option
              :label="$t('users.roleAdmin')"
              value="admin"
            />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreate = false">
          {{ $t('common.cancel') }}
        </el-button>
        <el-button
          type="primary"
          :loading="creating"
          @click="handleCreate"
        >
          {{ $t('common.create') }}
        </el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="showEdit"
      :title="$t('users.editTitle')"
      width="460px"
    >
      <el-form
        :model="editForm"
        label-width="80px"
      >
        <el-form-item :label="$t('users.name')">
          <el-input v-model="editForm.name" />
        </el-form-item>
        <el-form-item :label="$t('users.role')">
          <el-select
            v-model="editForm.role"
            style="width: 100%"
          >
            <el-option
              :label="$t('users.roleAgent')"
              value="agent"
            />
            <el-option
              :label="$t('users.roleAdmin')"
              value="admin"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('users.status')">
          <el-select
            v-model="editForm.status"
            style="width: 100%"
          >
            <el-option
              :label="$t('users.statusActive')"
              value="active"
            />
            <el-option
              :label="$t('users.statusDisabled')"
              value="disabled"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('users.passwordStatus')">
          <el-switch
            v-model="editForm.must_change_password"
            :active-text="$t('users.pwdNeedChange')"
            :inactive-text="$t('users.pwdSet')"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showEdit = false">
          {{ $t('common.cancel') }}
        </el-button>
        <el-button
          type="primary"
          :loading="saving"
          @click="handleUpdate"
        >
          {{ $t('common.save') }}
        </el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="showReset"
      :title="$t('users.resetTitle')"
      width="400px"
    >
      <div style="text-align: center; padding: 20px 0">
        <p style="margin-bottom: 12px; color: #909399">
          {{ $t('users.resetDesc') }}
        </p>
        <el-input
          :model-value="resetPasswordStr"
          readonly
          size="large"
          style="text-align: center; font-size: 20px; letter-spacing: 4px"
        >
          <template #append>
            <el-button @click="copyPassword">
              <el-icon><CopyDocument /></el-icon>
            </el-button>
          </template>
        </el-input>
        <p style="margin-top: 12px; color: #e6a23c; font-size: 13px">
          {{ $t('users.resetWarning') }}
        </p>
      </div>
      <template #footer>
        <el-button
          type="primary"
          @click="showReset = false"
        >
          {{ $t('common.close') }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { getUsers, createUser, updateUser, deleteUser, resetPassword } from '@/api/auth'
import { useUserStore } from '@/stores/user'

const { t, locale } = useI18n()
const userStore = useUserStore()
const users = ref([])
const loading = ref(false)
const creating = ref(false)
const saving = ref(false)

const showCreate = ref(false)
const showEdit = ref(false)
const showReset = ref(false)
const resetPasswordStr = ref('')

const createFormRef = ref<FormInstance>()
const createForm = reactive({ email: '', name: '', password: '', role: 'agent' })
const createRules: FormRules = {
  email: [
    { required: true, message: t('login.emailRequired'), trigger: 'blur' },
    { type: 'email', message: t('login.emailInvalid'), trigger: 'blur' },
  ],
  name: [{ required: true, message: t('init.nameRequired'), trigger: 'blur' }],
  password: [
    { required: true, message: t('init.passwordRequired'), trigger: 'blur' },
    { min: 6, message: t('init.passwordMin'), trigger: 'blur' },
  ],
  role: [{ required: true, message: 'Required', trigger: 'change' }],
}

const editForm = reactive({ id: '', name: '', role: '', status: '', must_change_password: false })

onMounted(() => fetchUsers())

async function fetchUsers() {
  loading.value = true
  try { const res = await getUsers(); users.value = res.data } finally { loading.value = false }
}

async function handleCreate() {
  const valid = await createFormRef.value?.validate().catch(() => false)
  if (!valid) return
  creating.value = true
  try {
    await createUser(createForm.email, createForm.name, createForm.password, createForm.role)
    ElMessage.success(t('users.createSuccess'))
    showCreate.value = false
    createForm.email = ''; createForm.name = ''; createForm.password = ''; createForm.role = 'agent'
    fetchUsers()
  } finally { creating.value = false }
}

function editUser(row: any) {
  editForm.id = row.id; editForm.name = row.name; editForm.role = row.role; editForm.status = row.status; editForm.must_change_password = row.must_change_password
  showEdit.value = true
}

async function handleUpdate() {
  saving.value = true
  try {
    await updateUser(editForm.id, { name: editForm.name, role: editForm.role, status: editForm.status, must_change_password: editForm.must_change_password })
    ElMessage.success(t('users.updateSuccess'))
    showEdit.value = false
    fetchUsers()
  } finally { saving.value = false }
}

async function handleReset(row: any) {
  try { await ElMessageBox.confirm(t('users.resetConfirm') + ' ' + row.email + '?', t('users.resetTitle'), { type: 'warning' }) } catch { return }
  try { const res = await resetPassword(row.id); resetPasswordStr.value = res.data.password; showReset.value = true; fetchUsers() } catch {}
}

function copyPassword() { navigator.clipboard.writeText(resetPasswordStr.value); ElMessage.success(t('common.copied')) }

async function handleDelete(row: any) {
  try { await ElMessageBox.confirm(t('users.deleteConfirm') + ' ' + row.email + '?', '', { type: 'warning' }) } catch { return }
  try { await deleteUser(row.id); ElMessage.success(t('common.success')); fetchUsers() } catch {}
}
</script>

<style scoped lang="scss">
.users-page { .card-header { display: flex; justify-content: space-between; align-items: center; } }
</style>
