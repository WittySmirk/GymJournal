// TODO: Create the modals
import "../App.css";
import { useState } from 'react';
import Modal from ".//modal";

function ExerciseButton(props: { create: boolean, name?: string, id?: string, fetchApi?: VoidFunction }) {
  const [modal, setModal] = useState<boolean>(false);
  async function modalAction(formData: any) {
    if (!props.create) {

      const weight = Number(formData.get("weight"));
      const sets = Number(formData.get("sets"));
      const reps = Number(formData.get("reps"));

      if (isNaN(weight) || isNaN(sets) || isNaN(reps)) {
        // TODO: Set up some type of error thing here that says must be a number
        setModal(false);
        return;
      }

      const d = {
        create: false,
        id: props.id,
        weight: weight,
        sets: sets,
        reps: reps,
      };

      await fetch("http://localhost:8080/app", {
        method: "POST",
        body: JSON.stringify(d),
        credentials: "include"
      });

      setModal(false);
      props.fetchApi!();
      return;
    }

    const name = String(formData.get("name"));
    if (!(typeof name === "string")) {
      // TODO: Set up some type of error thing here that says must be a number
      setModal(false);
      return;
    }

    const d = {
      create: true,
      name: name,
    };
    await fetch("http://localhost:8080/app", {
      method: "POST",
      body: JSON.stringify(d),
      credentials: "include"
    });

    setModal(false);
    props.fetchApi!();
  }

  return (
    <>
      <span onClick={() => setModal(true)} className="exerciseButton">
        {props.create ? "Create New" : props.name}
      </span>

      {modal ?
        <Modal formAction={modalAction} id={props.id} name={props.name}>
          {props.create ? (
            <>
              <h1>Create Exercise</h1>
              <label className="modal-item" htmlFor="name">Exercise Name</label>
              <input className="modal-item" type="text" id="name" name="name" required />
              <div className="modal-item">
                <input type="submit" value="submit" />
                <button type="button" onClick={() => setModal(false)}>Close</button>
              </div>
            </>
          ) : (
            <>
              <h1>{props.name}</h1>
              <label className="modal-item" htmlFor="weight">Weight</label>
              <input className="modal-item" type="text" id="weight" name="weight" required />
              <label className="modal-item" htmlFor="sets">Sets</label>
              <input className="modal-item" type="text" id="sets" name="sets" required />
              <label className="modal-item" htmlFor="reps">Reps</label>
              <input className="modal-item" type="text" id="reps" name="reps" required />
              <div className="modal-item">
                <input type="submit" value="submit" />
                <button type="button" onClick={() => setModal(false)}>Close</button>
              </div>

            </>
          )}

        </Modal >
        : <></>
      }
    </>
  );
}

export default ExerciseButton;
