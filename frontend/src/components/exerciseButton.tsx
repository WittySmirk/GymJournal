// TODO: Create the modals
import "../App.css";
import { useState } from 'react';
import Modal from "../components/modal.tsx"

function ExerciseButton(props: { create: boolean, name?: string, id?: string }) {
  const [modal, setModal] = useState<boolean>(false);
  /*
  async function postToServer() {
    const d = {
      create: props.create,
      id: props.id,
      message: "hello buddy"
    };
    // Not sure if we want to get the response or not
    await fetch("http://localhost:8080/app", {
      method: "POST",
      body: JSON.stringify(d)
    })
  }
  */

  // TODO: Figure out this type??

  return (
    <>
      <button onClick={() => setModal(true)} className="exerciseButton">
        {props.create ? "Create New" : props.name}
      </button>

      {modal ?

        <Modal create={props.create} id={props.id} name={props.name} resetModal={() => setModal(false)} />
        : <></>}
    </>
  );
}

export default ExerciseButton;
