import { useState, useEffect } from 'react'
import ExerciseButton from "./components/exerciseButton";
import './App.css'

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
  name: string,
  exercises: Exercise[]
};

function App() {
  const [data, setData] = useState<Data | undefined>(undefined);

  useEffect(() => {
    async function fetchApi() {
      const raw = await fetch("http://localhost:8080/app");
      const json = await raw.json();
      setData(json);
      console.log(data)
    }
    fetchApi();
  }, []);


  return (
    <>
      {data ? (
        <div className="exerciseButtonGrid">
          <h1>
            {data?.name}
          </h1>


          {data.exercises.map((e, k) => {
            return <ExerciseButton key={k} create={false} name={e.name} id={e.id} />
          })}
          <ExerciseButton create={true} />


        </div >

      ) : (
        <div>No data loaded :(</div>
      )}


    </>
  )
}

export default App
