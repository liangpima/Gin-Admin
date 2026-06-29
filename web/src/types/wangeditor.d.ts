declare module '@wangeditor/editor-for-vue' {
  import { DefineComponent } from 'vue'
  export const Editor: DefineComponent<any, any, any>
  export const Toolbar: DefineComponent<any, any, any>
}

declare module '@wangeditor/editor' {
  export interface IDomEditor {
    destroy(): void
  }
  export interface IToolbarConfig {}
  export interface IEditorConfig {
    placeholder?: string
  }
}
