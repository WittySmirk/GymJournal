import { useState, useEffect } from 'react'
import ExerciseButton from "../components/exerciseButton";
import Modal from "../components/modal";
import '../App.css'

/*
type Workout = {
  weight: number,
  sets: number,
  reps: number
};
*/

type Exercise = {
  id: string,
  name: string,
  //workouts: Workout[]
};

type Data = {
  nameNotValid: boolean,
  name: string,
  exercises: Exercise[]
};

function App() {
  const [data, setData] = useState<Data | undefined>(undefined);
  const [nameNotValid, setNameNotValid] = useState<boolean>(false);

  async function fetchApi() {
    // TODO: Figure out environment variables
    const raw = await fetch("http://localhost:8080/app", {
      credentials: "include",
    });


    const json = await raw.json();
    console.log(json);
    setData(json);
    if (data!.nameNotValid) {
      setNameNotValid(true);
    }
  }
  async function nameAction(formData: any) {
    const name = String(formData.get("name"));
    if (!(typeof name === "string")) {
      // TODO: set up error handling
      setNameNotValid(false);
      return;
    }
    const d = {
      createName: true,
      name: name,
    };

    await fetch("http://localhost:8080/app", {
      method: "POST",
      body: JSON.stringify(d),
      credentials: "include"
    });

    setNameNotValid(false);
    fetchApi();
  }

  useEffect(() => {
    fetchApi();
  }, []);

  return (
    <>
      {nameNotValid ? (
        <>
          <Modal formAction={nameAction} >
            <h1>Create User Name</h1>
            <label className="modal-item" htmlFor="name">Please Enter Your Name</label>
            <input className="modal-item" type="text" id="name" name="name" required />
            <div className="modal-item">
              <input type="submit" value="submit" />
            </div>
          </Modal>
        </>
      ) : (
        <>
        </>)}
      {data ? (
        <>
          <div className="navBar">
            <h1>
              {data?.name + "'s Gym Journal"}
            </h1>
            <a href="http://localhost:8080/logout/google">logout</a>
          </div>
          <div className="exerciseButtonGrid">
            {data.exercises.map((e, k) => {
              return <ExerciseButton key={k} create={false} name={e.name} id={e.id} />
            })}
            <ExerciseButton create={true} fetchApi={fetchApi} />
          </div >
        </>

      ) : (
        <div>No data loaded :(</div>
      )
      }


    </>
  )
}

export default App
