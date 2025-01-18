import { ReactNode } from "react";

function Modal(props: { children: ReactNode, formAction: (formData: any) => void, id?: string, name?: string }) {
  async function modalAction(formData: any) {
    props.formAction(formData);
  }
  return (
    <div id="modal">
      {/* @ts-ignore */}
      <form className="modal-form" action={modalAction}>
        {props.children}
      </form >
    </div >
  );
}

export default Modal;
